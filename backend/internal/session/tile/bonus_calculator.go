package tile

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/board"
	"terraforming-mars-backend/internal/session/deck"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// BonusCalculator calculates and awards bonuses from tile placement
type BonusCalculator struct {
	gameRepo   game.Repository
	playerRepo player.Repository
	boardRepo  board.Repository
	deckRepo   deck.Repository
	logger     *zap.Logger
}

// NewBonusCalculator creates a new bonus calculator
func NewBonusCalculator(gameRepo game.Repository, playerRepo player.Repository, boardRepo board.Repository, deckRepo deck.Repository) *BonusCalculator {
	return &BonusCalculator{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
		boardRepo:  boardRepo,
		deckRepo:   deckRepo,
		logger:     logger.Get(),
	}
}

// CalculateAndAwardBonuses calculates and awards all bonuses for a tile placement
func (bc *BonusCalculator) CalculateAndAwardBonuses(ctx context.Context, gameID, playerID string, coord board.HexPosition) error {
	log := bc.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("coordinate", coord.String()),
	)

	// Get tile to check for bonuses
	tile, err := bc.boardRepo.GetTile(ctx, gameID, coord)
	if err != nil {
		return fmt.Errorf("failed to get tile: %w", err)
	}

	// Award tile bonuses (steel, titanium, plants, card draw)
	if err := bc.awardTileBonuses(ctx, gameID, playerID, tile, log); err != nil {
		return fmt.Errorf("failed to award tile bonuses: %w", err)
	}

	// Award ocean adjacency bonus if applicable
	if err := bc.awardOceanAdjacencyBonus(ctx, gameID, playerID, coord, log); err != nil {
		return fmt.Errorf("failed to award ocean adjacency bonus: %w", err)
	}

	return nil
}

// awardTileBonuses awards bonuses from the tile itself (steel, titanium, plants, card draw)
func (bc *BonusCalculator) awardTileBonuses(ctx context.Context, gameID, playerID string, tile *board.Tile, log *zap.Logger) error {
	if len(tile.Bonuses) == 0 {
		return nil
	}

	p, err := bc.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	for _, bonus := range tile.Bonuses {
		switch bonus.Type {
		case board.ResourceSteel:
			p.Resources.Steel += bonus.Amount
			log.Info("🎁 Awarded steel from tile bonus",
				zap.Int("amount", bonus.Amount),
				zap.Int("new_total", p.Resources.Steel))

		case board.ResourceTitanium:
			p.Resources.Titanium += bonus.Amount
			log.Info("🎁 Awarded titanium from tile bonus",
				zap.Int("amount", bonus.Amount),
				zap.Int("new_total", p.Resources.Titanium))

		case board.ResourcePlants:
			p.Resources.Plants += bonus.Amount
			log.Info("🎁 Awarded plants from tile bonus",
				zap.Int("amount", bonus.Amount),
				zap.Int("new_total", p.Resources.Plants))

		case board.ResourceCardDraw:
			// Draw cards from deck
			drawnCards, err := bc.deckRepo.DrawProjectCards(ctx, gameID, bonus.Amount)
			if err != nil {
				return fmt.Errorf("failed to draw cards for tile bonus: %w", err)
			}

			// Create pending card draw selection (all cards are free from tile bonus)
			selection := &types.PendingCardDrawSelection{
				AvailableCards: drawnCards,
				FreeTakeCount:  bonus.Amount, // All cards must be taken (free from tile bonus)
				MaxBuyCount:    0,            // Cannot buy additional cards from tile bonus
				CardBuyCost:    0,
				Source:         "tile-bonus",
			}

			// Store pending selection for player to confirm
			if err := bc.playerRepo.UpdatePendingCardDrawSelection(ctx, gameID, playerID, selection); err != nil {
				return fmt.Errorf("failed to create pending card draw selection: %w", err)
			}

			log.Info("🎁 Awarded card draw bonus from tile",
				zap.Int("amount", bonus.Amount),
				zap.Strings("cards", drawnCards))

		default:
			log.Warn("⚠️  Unknown bonus type encountered",
				zap.String("bonus_type", string(bonus.Type)))
		}
	}

	// Update player resources
	if err := bc.playerRepo.UpdateResources(ctx, gameID, playerID, p.Resources); err != nil {
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	return nil
}

// awardOceanAdjacencyBonus awards megacredits for placing tiles adjacent to oceans
func (bc *BonusCalculator) awardOceanAdjacencyBonus(ctx context.Context, gameID, playerID string, coord board.HexPosition, log *zap.Logger) error {
	// Get board to check for adjacent oceans
	b, err := bc.boardRepo.GetByGameID(ctx, gameID)
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
	p, err := bc.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	hasLakefrontResorts := bc.hasCard(p, "061")
	bonusPerOcean := 2
	if hasLakefrontResorts {
		bonusPerOcean = 4 // 2 base + 2 from Lakefront Resorts
	}

	totalBonus := bonusPerOcean * adjacentOceans

	// Award megacredits
	p.Resources.Credits += totalBonus
	if err := bc.playerRepo.UpdateResources(ctx, gameID, playerID, p.Resources); err != nil {
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	if hasLakefrontResorts {
		log.Info("🎁 Awarded ocean adjacency bonus with Lakefront Resorts",
			zap.Int("adjacent_oceans", adjacentOceans),
			zap.Int("bonus_per_ocean", bonusPerOcean),
			zap.Int("total_bonus", totalBonus),
			zap.Int("new_credits", p.Resources.Credits))
	} else {
		log.Info("🎁 Awarded ocean adjacency bonus",
			zap.Int("adjacent_oceans", adjacentOceans),
			zap.Int("bonus_per_ocean", bonusPerOcean),
			zap.Int("total_bonus", totalBonus),
			zap.Int("new_credits", p.Resources.Credits))
	}

	return nil
}

// countAdjacentOceans counts how many ocean tiles are adjacent to the given coordinate
func (bc *BonusCalculator) countAdjacentOceans(coord board.HexPosition, b *board.Board) int {
	// Cube coordinate neighbors (Q, R, S)
	neighbors := []board.HexPosition{
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
				if tile.Type == board.ResourceOceanTile {
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
	for _, playedCardID := range p.PlayedCards {
		if playedCardID == cardID {
			return true
		}
	}
	return false
}
