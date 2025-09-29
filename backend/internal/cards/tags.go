package cards

import (
	"context"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// TagManager handles all tag-related operations and utilities
type TagManager struct {
	cardRepo repository.CardRepository
}

// NewTagManager creates a new tag manager
func NewTagManager(cardRepo repository.CardRepository) *TagManager {
	return &TagManager{
		cardRepo: cardRepo,
	}
}

// CountPlayerTags counts the occurrence of each tag in player's played cards and corporation
func (tm *TagManager) CountPlayerTags(ctx context.Context, player *model.Player) map[model.CardTag]int {
	tagCounts := make(map[model.CardTag]int)

	// Count tags from played cards
	for _, cardID := range player.PlayedCards {
		card, err := tm.cardRepo.GetCardByID(ctx, cardID)
		if err != nil || card == nil {
			continue // Skip if card not found
		}

		for _, tag := range card.Tags {
			tagCounts[tag]++
		}
	}

	// Add corporation tags if player has a corporation
	if player.Corporation != nil && *player.Corporation != "" {
		corporationCard, err := tm.cardRepo.GetCardByID(ctx, *player.Corporation)
		if err == nil && corporationCard != nil {
			for _, tag := range corporationCard.Tags {
				tagCounts[tag]++
			}
		}
	}

	return tagCounts
}

// GetPlayerTagCount returns the count of a specific tag for a player
func (tm *TagManager) GetPlayerTagCount(ctx context.Context, player *model.Player, tag model.CardTag) int {
	tagCounts := tm.CountPlayerTags(ctx, player)
	return tagCounts[tag]
}
