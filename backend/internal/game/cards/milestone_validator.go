package cards

import (
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// MilestoneType represents the type of milestone (duplicated to avoid import cycle)
type MilestoneType string

const (
	MilestoneTerraformer MilestoneType = "terraformer"
	MilestoneMayor       MilestoneType = "mayor"
	MilestoneGardener    MilestoneType = "gardener"
	MilestoneBuilder     MilestoneType = "builder"
	MilestonePlanner     MilestoneType = "planner"
)

// MilestoneRequirement holds the requirement threshold for a milestone
type MilestoneRequirement struct {
	Description string
	Required    int
}

// GetMilestoneRequirement returns the requirement for a specific milestone type
func GetMilestoneRequirement(milestoneType MilestoneType) MilestoneRequirement {
	switch milestoneType {
	case MilestoneTerraformer:
		return MilestoneRequirement{
			Description: "Terraform Rating of at least 35",
			Required:    35,
		}
	case MilestoneMayor:
		return MilestoneRequirement{
			Description: "Own at least 3 city tiles",
			Required:    3,
		}
	case MilestoneGardener:
		return MilestoneRequirement{
			Description: "Own at least 3 greenery tiles",
			Required:    3,
		}
	case MilestoneBuilder:
		return MilestoneRequirement{
			Description: "Have at least 8 building tags in play",
			Required:    8,
		}
	case MilestonePlanner:
		return MilestoneRequirement{
			Description: "Have at least 16 cards in hand",
			Required:    16,
		}
	default:
		return MilestoneRequirement{}
	}
}

// CardRegistryInterface defines the interface for looking up cards
// This avoids an import cycle with the cards package
type CardRegistryInterface interface {
	GetByID(cardID string) (*Card, error)
}

// CanClaimMilestone checks if a player meets the requirements for a milestone
func CanClaimMilestone(
	milestoneType MilestoneType,
	p *player.Player,
	b *board.Board,
	cardRegistry CardRegistryInterface,
) bool {
	current := GetPlayerMilestoneProgress(milestoneType, p, b, cardRegistry)
	required := GetMilestoneRequirement(milestoneType).Required
	return current >= required
}

// GetPlayerMilestoneProgress returns the current progress for a player towards a milestone
func GetPlayerMilestoneProgress(
	milestoneType MilestoneType,
	p *player.Player,
	b *board.Board,
	cardRegistry CardRegistryInterface,
) int {
	switch milestoneType {
	case MilestoneTerraformer:
		return p.Resources().TerraformRating()
	case MilestoneMayor:
		return countPlayerTiles(p.ID(), b, shared.ResourceCityTile)
	case MilestoneGardener:
		return countPlayerTiles(p.ID(), b, shared.ResourceGreeneryTile)
	case MilestoneBuilder:
		return countPlayerBuildingTags(p, cardRegistry)
	case MilestonePlanner:
		return p.Hand().CardCount()
	default:
		return 0
	}
}

// countPlayerTiles counts tiles of a specific type owned by a player
func countPlayerTiles(playerID string, b *board.Board, tileType shared.ResourceType) int {
	count := 0
	tiles := b.Tiles()
	for _, tile := range tiles {
		if tile.OwnerID != nil && *tile.OwnerID == playerID {
			if tile.OccupiedBy != nil && tile.OccupiedBy.Type == tileType {
				count++
			}
		}
	}
	return count
}

// countPlayerBuildingTags counts building tags across all played cards
func countPlayerBuildingTags(p *player.Player, cardRegistry CardRegistryInterface) int {
	count := 0
	playedCardIDs := p.PlayedCards().Cards()

	for _, cardID := range playedCardIDs {
		card, err := cardRegistry.GetByID(cardID)
		if err != nil {
			continue // Skip cards not in registry
		}
		for _, tag := range card.Tags {
			if tag == shared.TagBuilding {
				count++
			}
		}
	}

	return count
}
