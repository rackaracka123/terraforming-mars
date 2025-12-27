package cards

import (
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// MilestoneRequirement holds the requirement threshold for a milestone
type MilestoneRequirement struct {
	Description string
	Required    int
}

// GetMilestoneRequirement returns the requirement for a specific milestone type
func GetMilestoneRequirement(milestoneType shared.MilestoneType) MilestoneRequirement {
	switch milestoneType {
	case shared.MilestoneTerraformer:
		return MilestoneRequirement{
			Description: "Terraform Rating of at least 35",
			Required:    35,
		}
	case shared.MilestoneMayor:
		return MilestoneRequirement{
			Description: "Own at least 3 city tiles",
			Required:    3,
		}
	case shared.MilestoneGardener:
		return MilestoneRequirement{
			Description: "Own at least 3 greenery tiles",
			Required:    3,
		}
	case shared.MilestoneBuilder:
		return MilestoneRequirement{
			Description: "Have at least 8 building tags in play",
			Required:    8,
		}
	case shared.MilestonePlanner:
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
	milestoneType shared.MilestoneType,
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
	milestoneType shared.MilestoneType,
	p *player.Player,
	b *board.Board,
	cardRegistry CardRegistryInterface,
) int {
	cityTileType := shared.ResourceCityTile
	greeneryTileType := shared.ResourceGreeneryTile

	switch milestoneType {
	case shared.MilestoneTerraformer:
		return p.Resources().TerraformRating()
	case shared.MilestoneMayor:
		return CountPlayerTiles(p.ID(), b, &cityTileType)
	case shared.MilestoneGardener:
		return CountPlayerTiles(p.ID(), b, &greeneryTileType)
	case shared.MilestoneBuilder:
		return CountPlayerTagsByType(p, cardRegistry, shared.TagBuilding)
	case shared.MilestonePlanner:
		return p.Hand().CardCount()
	default:
		return 0
	}
}
