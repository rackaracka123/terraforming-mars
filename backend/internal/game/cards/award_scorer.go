package cards

import (
	"sort"

	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// AwardType represents the type of award (duplicated to avoid import cycle)
type AwardType string

const (
	AwardLandlord   AwardType = "landlord"
	AwardBanker     AwardType = "banker"
	AwardScientist  AwardType = "scientist"
	AwardThermalist AwardType = "thermalist"
	AwardMiner      AwardType = "miner"
)

// AwardPlacement represents a player's placement in an award
type AwardPlacement struct {
	PlayerID  string
	Score     int
	Placement int // 1 = first place (5 VP), 2 = second place (2 VP), 0 = no placement
}

// CalculateAwardScore calculates a player's score for a specific award
func CalculateAwardScore(
	awardType AwardType,
	p *player.Player,
	b *board.Board,
	cardRegistry CardRegistryInterface,
) int {
	switch awardType {
	case AwardLandlord:
		// Most tiles on Mars (owned by player)
		return countAllPlayerTiles(p.ID(), b)
	case AwardBanker:
		// Highest MC production
		production := p.Resources().Production()
		return production.Credits
	case AwardScientist:
		// Most science tags in play
		return countPlayerTagsByType(p, cardRegistry, shared.TagScience)
	case AwardThermalist:
		// Most heat resources
		resources := p.Resources().Get()
		return resources.Heat
	case AwardMiner:
		// Most steel + titanium resources
		resources := p.Resources().Get()
		return resources.Steel + resources.Titanium
	default:
		return 0
	}
}

// ScoreAward calculates placements for all players for an award
// Returns a slice of AwardPlacement sorted by placement (1st, 2nd, then others)
func ScoreAward(
	awardType AwardType,
	players []*player.Player,
	b *board.Board,
	cardRegistry CardRegistryInterface,
) []AwardPlacement {
	// Calculate scores for all players
	placements := make([]AwardPlacement, len(players))
	for i, p := range players {
		placements[i] = AwardPlacement{
			PlayerID: p.ID(),
			Score:    CalculateAwardScore(awardType, p, b, cardRegistry),
		}
	}

	// Sort by score descending
	sort.Slice(placements, func(i, j int) bool {
		return placements[i].Score > placements[j].Score
	})

	// Assign placements (handle ties)
	if len(placements) == 0 {
		return placements
	}

	// First place(s)
	firstPlaceScore := placements[0].Score
	for i := range placements {
		if placements[i].Score == firstPlaceScore {
			placements[i].Placement = 1
		} else {
			break
		}
	}

	// Count how many tied for first
	firstPlaceCount := 0
	for _, p := range placements {
		if p.Placement == 1 {
			firstPlaceCount++
		}
	}

	// Second place(s) - only if not everyone tied for first
	if firstPlaceCount < len(placements) {
		// Find the next highest score
		var secondPlaceScore int
		foundSecond := false
		for _, p := range placements {
			if p.Placement != 1 {
				secondPlaceScore = p.Score
				foundSecond = true
				break
			}
		}

		if foundSecond {
			for i := range placements {
				if placements[i].Placement == 0 && placements[i].Score == secondPlaceScore {
					placements[i].Placement = 2
				}
			}
		}
	}

	return placements
}

// GetAwardVP returns the VP for a specific placement
func GetAwardVP(placement int) int {
	switch placement {
	case 1:
		return 5
	case 2:
		return 2
	default:
		return 0
	}
}

// countAllPlayerTiles counts all tiles (city, greenery, etc.) owned by a player
func countAllPlayerTiles(playerID string, b *board.Board) int {
	count := 0
	tiles := b.Tiles()
	for _, tile := range tiles {
		if tile.OwnerID != nil && *tile.OwnerID == playerID {
			if tile.OccupiedBy != nil {
				count++
			}
		}
	}
	return count
}

// countPlayerTagsByType counts tags of a specific type across all played cards
func countPlayerTagsByType(p *player.Player, cardRegistry CardRegistryInterface, tagType shared.CardTag) int {
	count := 0
	playedCardIDs := p.PlayedCards().Cards()

	for _, cardID := range playedCardIDs {
		card, err := cardRegistry.GetByID(cardID)
		if err != nil {
			continue // Skip cards not in registry
		}
		for _, tag := range card.Tags {
			if tag == tagType {
				count++
			}
		}
	}

	return count
}
