package board

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/game/deck"
	"terraforming-mars-backend/internal/session/game/player"

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
// Note: gameID is not needed as this calculator is scoped to a specific game instance
func (bc *BonusCalculator) CalculateAndAwardBonuses(ctx context.Context, p *player.Player, coord HexPosition) error {
	log := bc.logger.With(
		zap.String("player_id", p.ID()),
		zap.String("coordinate", coord.String()),
	)

	// Get tile to check for bonuses
	tile, err := bc.boardRepo.GetTile(ctx, coord)
	if err != nil {
		return fmt.Errorf("failed to get tile: %w", err)
	}

	// Award tile bonuses (steel, titanium, plants, card draw)
	if err := bc.awardTileBonuses(ctx, p, tile, log); err != nil {
		return fmt.Errorf("failed to award tile bonuses: %w", err)
	}

	// Award ocean adjacency bonus if applicable
	if err := bc.awardOceanAdjacencyBonus(ctx, p, coord, log); err != nil {
		return fmt.Errorf("failed to award ocean adjacency bonus: %w", err)
	}

	return nil
}

// awardTileBonuses awards bonuses from the tile itself (steel, titanium, plants, card draw)
func (bc *BonusCalculator) awardTileBonuses(ctx context.Context, p *player.Player, tile *Tile, log *zap.Logger) error {
	if len(tile.Bonuses) == 0 {
		return nil
	}

	// Get current resources
	resources := p.Resources()

	for _, bonus := range tile.Bonuses {
		switch bonus.Type {
		case ResourceSteel:
			resources.Steel += bonus.Amount
			log.Info("🎁 Awarded steel from tile bonus",
				zap.Int("amount", bonus.Amount),
				zap.Int("new_total", resources.Steel))

		case ResourceTitanium:
			resources.Titanium += bonus.Amount
			log.Info("🎁 Awarded titanium from tile bonus",
				zap.Int("amount", bonus.Amount),
				zap.Int("new_total", resources.Titanium))

		case ResourcePlants:
			resources.Plants += bonus.Amount
			log.Info("🎁 Awarded plants from tile bonus",
				zap.Int("amount", bonus.Amount),
				zap.Int("new_total", resources.Plants))

		case ResourceCardDraw:
			// Draw cards from deck
			drawnCards, err := bc.deckRepo.DrawProjectCards(ctx, bonus.Amount)
			if err != nil {
				return fmt.Errorf("failed to draw cards for tile bonus: %w", err)
			}

			// Create pending card draw selection (all cards are free from tile bonus)
			selection := &player.PendingCardDrawSelection{
				AvailableCards: drawnCards,
				FreeTakeCount:  bonus.Amount, // All cards must be taken (free from tile bonus)
				MaxBuyCount:    0,            // Cannot buy additional cards from tile bonus
				CardBuyCost:    0,
				Source:         "tile-bonus",
			}

			// Store pending selection using player setter
			if err := p.SetPendingCardDrawSelection(ctx, selection); err != nil {
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

	// Update resources using player setter
	if err := p.SetResources(ctx, resources); err != nil {
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	return nil
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

	// Award megacredits using player setter
	resources := p.Resources()
	resources.Credits += totalBonus

	if err := p.SetResources(ctx, resources); err != nil {
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	if hasLakefrontResorts {
		log.Info("🎁 Awarded ocean adjacency bonus with Lakefront Resorts",
			zap.Int("adjacent_oceans", adjacentOceans),
			zap.Int("bonus_per_ocean", bonusPerOcean),
			zap.Int("total_bonus", totalBonus),
			zap.Int("new_credits", resources.Credits))
	} else {
		log.Info("🎁 Awarded ocean adjacency bonus",
			zap.Int("adjacent_oceans", adjacentOceans),
			zap.Int("bonus_per_ocean", bonusPerOcean),
			zap.Int("total_bonus", totalBonus),
			zap.Int("new_credits", resources.Credits))
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
	playedCards := p.PlayedCards()
	for _, playedCardID := range playedCards {
		if playedCardID == cardID {
			return true
		}
	}
	return false
}
