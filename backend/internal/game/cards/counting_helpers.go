package cards

import (
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// CountPlayerTiles counts tiles owned by a player on the board.
// If tileType is nil, counts all tiles owned by the player.
// If tileType is specified, counts only tiles of that type.
func CountPlayerTiles(playerID string, b *board.Board, tileType *shared.ResourceType) int {
	count := 0
	tiles := b.Tiles()
	for _, tile := range tiles {
		if tile.OwnerID == nil || *tile.OwnerID != playerID {
			continue
		}
		if tile.OccupiedBy == nil {
			continue
		}
		if tileType != nil && tile.OccupiedBy.Type != *tileType {
			continue
		}
		count++
	}
	return count
}

// CountPlayerTagsByType counts tags of a specific type across all played cards for a player.
func CountPlayerTagsByType(p *player.Player, cardRegistry CardRegistryInterface, tagType shared.CardTag) int {
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

// CountAllTilesOfType counts all tiles of a specific type on the board, regardless of owner.
func CountAllTilesOfType(b *board.Board, tileType shared.ResourceType) int {
	count := 0
	tiles := b.Tiles()
	for _, tile := range tiles {
		if tile.OccupiedBy != nil && tile.OccupiedBy.Type == tileType {
			count++
		}
	}
	return count
}
