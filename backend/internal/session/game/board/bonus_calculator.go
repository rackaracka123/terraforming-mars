package board

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/game/deck"
	"terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/game/player/selection"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// BonusCalculator calculates and awards bonuses from tile placement
type BonusCalculator struct {
	boardRepo Repository
	deckRepo  deck.Repository
	logger    *zap.Logger
}

// NewBonusCalculator creates a new bonus calculator
func NewBonusCalculator(boardRepo Repository, deckRepo deck.Repository) *BonusCalculator {
	return &BonusCalculator{
		boardRepo: boardRepo,
		deckRepo:  deckRepo,
		logger:    logger.Get(),
	}
}

// CalculateAndAwardBonuses calculates and awards all bonuses for a tile placement
// Returns any pending card draw selection that needs to be set on Player
func (bc *BonusCalculator) CalculateAndAwardBonuses(ctx context.Context, p *player.Player, coord HexPosition) (*selection.PendingCardDrawSelection, error) {
	log := bc.logger.With(
		zap.String("player_id", p.ID()),
		zap.String("coordinate", coord.String()),
	)

	// Get tile to check for bonuses
	tile, err := bc.boardRepo.GetTile(ctx, coord)
	if err != nil {
		return nil, fmt.Errorf("failed to get tile: %w", err)
	}

	// Award tile bonuses (steel, titanium, plants, card draw)
	pendingSelection, err := bc.awardTileBonuses(ctx, p, tile, log)
	if err != nil {
		return nil, fmt.Errorf("failed to award tile bonuses: %w", err)
	}

	// Award ocean adjacency bonus if applicable
	if err := bc.awardOceanAdjacencyBonus(ctx, p, coord, log); err != nil {
		return nil, fmt.Errorf("failed to award ocean adjacency bonus: %w", err)
	}

	return pendingSelection, nil
}

// awardTileBonuses awards bonuses from the tile itself (steel, titanium, plants, card draw)
// Returns any pending card draw selection that needs to be set on Player
func (bc *BonusCalculator) awardTileBonuses(ctx context.Context, p *player.Player, tile *Tile, log *zap.Logger) (*selection.PendingCardDrawSelection, error) {
	if len(tile.Bonuses) == 0 {
		return nil, nil
	}

	// Collect resource changes for batched update
	resourceChanges := make(map[types.ResourceType]int)
	var pendingSelection *selection.PendingCardDrawSelection

	for _, bonus := range tile.Bonuses {
		switch bonus.Type {
		case ResourceSteel:
			resourceChanges[types.ResourceSteel] += bonus.Amount
			currentSteel := p.Resources().Get().Steel + resourceChanges[types.ResourceSteel]
			log.Info("🎁 Awarded steel from tile bonus",
				zap.Int("amount", bonus.Amount),
				zap.Int("new_total", currentSteel))

		case ResourceTitanium:
			resourceChanges[types.ResourceTitanium] += bonus.Amount
			currentTitanium := p.Resources().Get().Titanium + resourceChanges[types.ResourceTitanium]
			log.Info("🎁 Awarded titanium from tile bonus",
				zap.Int("amount", bonus.Amount),
				zap.Int("new_total", currentTitanium))

		case ResourcePlants:
			resourceChanges[types.ResourcePlants] += bonus.Amount
			currentPlants := p.Resources().Get().Plants + resourceChanges[types.ResourcePlants]
			log.Info("🎁 Awarded plants from tile bonus",
				zap.Int("amount", bonus.Amount),
				zap.Int("new_total", currentPlants))

		case ResourceCardDraw:
			// Draw cards from deck
			drawnCards, err := bc.deckRepo.DrawProjectCards(ctx, bonus.Amount)
			if err != nil {
				return nil, fmt.Errorf("failed to draw cards for tile bonus: %w", err)
			}

			// Create pending card draw selection (all cards are free from tile bonus)
			// This will be set on Player by the caller
			pendingSelection = &selection.PendingCardDrawSelection{
				AvailableCards: drawnCards,
				FreeTakeCount:  bonus.Amount, // All cards must be taken (free from tile bonus)
				MaxBuyCount:    0,            // Cannot buy additional cards from tile bonus
				CardBuyCost:    0,
				Source:         "tile-bonus",
			}

			log.Info("🎁 Awarded card draw bonus from tile",
				zap.Int("amount", bonus.Amount),
				zap.Strings("cards", drawnCards))

		default:
			log.Warn("⚠️  Unknown bonus type encountered",
				zap.String("bonus_type", string(bonus.Type)))
		}
	}

	// Apply resource changes using component's Add method (publishes events automatically)
	if len(resourceChanges) > 0 {
		p.Resources().Add(resourceChanges)
	}

	return pendingSelection, nil
}

// awardOceanAdjacencyBonus awards megacredits for placing tiles adjacent to oceans
func (bc *BonusCalculator) awardOceanAdjacencyBonus(ctx context.Context, p *player.Player, coord HexPosition, log *zap.Logger) error {
	// Get board to check for adjacent oceans
	b, err := bc.boardRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get board: %w", err)
	}

	// Count adjacent ocean tiles
	adjacentOceans := bc.countAdjacentOceans(coord, b)
	if adjacentOceans == 0 {
		return nil
	}

	// Check for Lakefront Resorts card effect (+2 MC per ocean, total 4 MC)
	// Card ID: 061 - Lakefront Resorts
	hasLakefrontResorts := bc.hasCard(p, "061")
	bonusPerOcean := 2
	if hasLakefrontResorts {
		bonusPerOcean = 4 // 2 base + 2 from Lakefront Resorts
	}

	totalBonus := bonusPerOcean * adjacentOceans

	// Award megacredits using component's Add method (publishes events automatically)
	p.Resources().Add(map[types.ResourceType]int{
		types.ResourceCredits: totalBonus,
	})

	newCredits := p.Resources().Get().Credits

	if hasLakefrontResorts {
		log.Info("🎁 Awarded ocean adjacency bonus with Lakefront Resorts",
			zap.Int("adjacent_oceans", adjacentOceans),
			zap.Int("bonus_per_ocean", bonusPerOcean),
			zap.Int("total_bonus", totalBonus),
			zap.Int("new_credits", newCredits))
	} else {
		log.Info("🎁 Awarded ocean adjacency bonus",
			zap.Int("adjacent_oceans", adjacentOceans),
			zap.Int("bonus_per_ocean", bonusPerOcean),
			zap.Int("total_bonus", totalBonus),
			zap.Int("new_credits", newCredits))
	}

	return nil
}

// countAdjacentOceans counts how many ocean tiles are adjacent to the given coordinate
func (bc *BonusCalculator) countAdjacentOceans(coord HexPosition, b *Board) int {
	// Cube coordinate neighbors (Q, R, S)
	neighbors := []HexPosition{
		{Q: coord.Q + 1, R: coord.R - 1, S: coord.S}, // East
		{Q: coord.Q + 1, R: coord.R, S: coord.S - 1}, // Southeast
		{Q: coord.Q, R: coord.R + 1, S: coord.S - 1}, // Southwest
		{Q: coord.Q - 1, R: coord.R + 1, S: coord.S}, // West
		{Q: coord.Q - 1, R: coord.R, S: coord.S + 1}, // Northwest
		{Q: coord.Q, R: coord.R - 1, S: coord.S + 1}, // Northeast
	}

	oceanCount := 0
	for _, neighborCoord := range neighbors {
		// Find tile in board
		for _, tile := range b.Tiles {
			if tile.Coordinates.Q == neighborCoord.Q &&
				tile.Coordinates.R == neighborCoord.R &&
				tile.Coordinates.S == neighborCoord.S {
				// Check if tile is an ocean
				if tile.Type == ResourceOceanTile {
					oceanCount++
				}
				break
			}
		}
	}

	return oceanCount
}

// hasCard checks if player has a specific card by ID
func (bc *BonusCalculator) hasCard(p *player.Player, cardID string) bool {
	playedCards := p.Hand().PlayedCards()
	for _, playedCardID := range playedCards {
		if playedCardID == cardID {
			return true
		}
	}
	return false
}
