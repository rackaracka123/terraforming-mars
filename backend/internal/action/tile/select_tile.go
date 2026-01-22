package tile

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	baseaction "terraforming-mars-backend/internal/action"
	"time"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/shared"
)

// SelectTileAction handles the business logic for selecting a tile position
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type SelectTileAction struct {
	baseaction.BaseAction
}

// NewSelectTileAction creates a new select tile action
func NewSelectTileAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *SelectTileAction {
	return &SelectTileAction{
		BaseAction: baseaction.NewBaseAction(gameRepo, cardRegistry),
	}
}

// Execute performs the select tile action
func (a *SelectTileAction) Execute(ctx context.Context, gameID string, playerID string, selectedHex string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "select_tile"))
	log.Info("🎯 Selecting tile", zap.String("hex", selectedHex))

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	pendingTileSelection := g.GetPendingTileSelection(playerID)
	if pendingTileSelection == nil {
		log.Warn("No pending tile selection found")
		return fmt.Errorf("no pending tile selection found for player %s", playerID)
	}

	hexIsValid := false
	for _, availableHex := range pendingTileSelection.AvailableHexes {
		if availableHex == selectedHex {
			hexIsValid = true
			break
		}
	}
	if !hexIsValid {
		log.Warn("Invalid hex selection",
			zap.String("selected_hex", selectedHex),
			zap.Strings("available_hexes", pendingTileSelection.AvailableHexes))
		return fmt.Errorf("selected hex %s is not valid for placement", selectedHex)
	}

	coords, err := parseHexPosition(selectedHex)
	if err != nil {
		log.Warn("Failed to parse hex coordinates", zap.String("hex", selectedHex), zap.Error(err))
		return fmt.Errorf("invalid hex format: %w", err)
	}

	tileType := pendingTileSelection.TileType
	occupant := board.TileOccupant{
		Type: mapTileTypeToResourceType(tileType),
		Tags: []string{},
	}

	if err := g.Board().UpdateTileOccupancy(ctx, *coords, occupant, playerID); err != nil {
		log.Warn("Failed to place tile", zap.Error(err))
		return fmt.Errorf("failed to place tile: %w", err)
	}

	log.Info("🏗️ Tile placed on board",
		zap.String("tile_type", tileType),
		zap.String("position", selectedHex))

	placedTile, err := g.Board().GetTile(*coords)
	if err != nil {
		log.Warn("Failed to get placed tile for bonus check", zap.Error(err))
		// Continue execution - tile was placed successfully, just couldn't check for bonuses
	} else if len(placedTile.Bonuses) > 0 {
		log.Info("🎁 Tile has bonuses", zap.Int("bonus_count", len(placedTile.Bonuses)))

		resourceBonuses := make(map[string]int)

		for _, bonus := range placedTile.Bonuses {
			switch bonus.Type {
			case shared.ResourceSteel, shared.ResourceTitanium, shared.ResourcePlant:
				player.Resources().Add(map[shared.ResourceType]int{
					bonus.Type: bonus.Amount,
				})
				log.Info("💰 Awarded resource bonus",
					zap.String("resource", string(bonus.Type)),
					zap.Int("amount", bonus.Amount))

				resourceBonuses[string(bonus.Type)] = bonus.Amount

			case shared.ResourceCardDraw:
				deck := g.Deck()
				cardIDs, err := deck.DrawProjectCards(ctx, bonus.Amount)
				if err != nil {
					log.Warn("Failed to draw cards for bonus", zap.Error(err))
					// Continue - other bonuses may still be valid
					continue
				}

				for _, cardID := range cardIDs {
					player.Hand().AddCard(cardID)
				}

				log.Info("🃏 Awarded card draw bonus",
					zap.Int("cards_drawn", len(cardIDs)),
					zap.Strings("card_ids", cardIDs))

			default:
				log.Warn("⚠️  Unhandled tile bonus type",
					zap.String("type", string(bonus.Type)),
					zap.Int("amount", bonus.Amount))
			}
		}

		if len(resourceBonuses) > 0 {
			events.Publish(g.EventBus(), events.PlacementBonusGainedEvent{
				GameID:    gameID,
				PlayerID:  playerID,
				Resources: resourceBonuses,
				Q:         coords.Q,
				R:         coords.R,
				S:         coords.S,
				Timestamp: time.Now(),
			})
			log.Info("📢 Published PlacementBonusGainedEvent",
				zap.Any("resources", resourceBonuses))
		}
	}

	switch tileType {
	case "city":
		log.Info("🏙️ City placed (no TR bonus)")

	case "greenery":
		actualSteps, err := g.GlobalParameters().IncreaseOxygen(ctx, 1)
		if err != nil {
			return fmt.Errorf("failed to increase oxygen: %w", err)
		}
		if actualSteps > 0 {
			currentTR := player.Resources().TerraformRating()
			player.Resources().UpdateTerraformRating(1)
			log.Info("🌿 Increased oxygen and TR for greenery placement",
				zap.Int("oxygen_steps", actualSteps),
				zap.Int("new_tr", currentTR+1))
		} else {
			log.Info("🌿 Greenery placed but oxygen already maxed")
		}

	case "ocean":
		success, err := g.GlobalParameters().PlaceOcean(ctx)
		if err != nil {
			return fmt.Errorf("failed to place ocean: %w", err)
		}
		if success {
			currentTR := player.Resources().TerraformRating()
			player.Resources().UpdateTerraformRating(1)
			log.Info("🌊 Placed ocean and increased TR",
				zap.Int("new_tr", currentTR+1))
		} else {
			log.Info("🌊 Ocean placed but ocean count already maxed")
		}
	}

	if err := g.SetPendingTileSelection(ctx, playerID, nil); err != nil {
		return fmt.Errorf("failed to clear pending tile selection: %w", err)
	}

	if err := g.ProcessNextTile(ctx, playerID); err != nil {
		return fmt.Errorf("failed to process next tile: %w", err)
	}

	log.Info("✅ Tile selected and placed successfully",
		zap.String("tile_type", tileType),
		zap.String("position", selectedHex))
	return nil
}

// parseHexPosition parses a hex position string in the format "q,r,s"
func parseHexPosition(hexStr string) (*shared.HexPosition, error) {
	parts := strings.Split(hexStr, ",")
	if len(parts) != 3 {
		return nil, fmt.Errorf("expected 3 coordinates, got %d", len(parts))
	}

	q, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid q coordinate: %w", err)
	}

	r, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid r coordinate: %w", err)
	}

	s, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid s coordinate: %w", err)
	}

	// Validate cube coordinates constraint: q + r + s = 0
	if q+r+s != 0 {
		return nil, fmt.Errorf("invalid cube coordinates: q+r+s must equal 0")
	}

	return &shared.HexPosition{Q: q, R: r, S: s}, nil
}

// mapTileTypeToResourceType maps a tile type string to a ResourceType
func mapTileTypeToResourceType(tileType string) shared.ResourceType {
	switch tileType {
	case "city":
		return shared.ResourceCityTile
	case "greenery":
		return shared.ResourceGreeneryTile
	case "ocean":
		return shared.ResourceOceanTile
	default:
		return shared.ResourceType(tileType)
	}
}
