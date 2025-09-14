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

// HasRequiredTags checks if a player has all the required tags
func (tm *TagManager) HasRequiredTags(ctx context.Context, player *model.Player, requiredTags []model.CardTag) bool {
	if len(requiredTags) == 0 {
		return true
	}

	playerTagCounts := tm.CountPlayerTags(ctx, player)

	for _, requiredTag := range requiredTags {
		if playerTagCounts[requiredTag] == 0 {
			return false
		}
	}

	return true
}

// GetMissingTags returns a list of tags that the player doesn't have but are required
func (tm *TagManager) GetMissingTags(ctx context.Context, player *model.Player, requiredTags []model.CardTag) []model.CardTag {
	if len(requiredTags) == 0 {
		return nil
	}

	playerTagCounts := tm.CountPlayerTags(ctx, player)
	var missingTags []model.CardTag

	for _, requiredTag := range requiredTags {
		if playerTagCounts[requiredTag] == 0 {
			missingTags = append(missingTags, requiredTag)
		}
	}

	return missingTags
}

// GetTagsFromCard returns all tags from a specific card
func (tm *TagManager) GetTagsFromCard(ctx context.Context, cardID string) ([]model.CardTag, error) {
	card, err := tm.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if card == nil {
		return nil, nil
	}

	return card.Tags, nil
}

// GetPlayersWithTag returns a list of players who have the specified tag
func (tm *TagManager) GetPlayersWithTag(ctx context.Context, players []model.Player, tag model.CardTag) []model.Player {
	var playersWithTag []model.Player

	for _, player := range players {
		if tm.GetPlayerTagCount(ctx, &player, tag) > 0 {
			playersWithTag = append(playersWithTag, player)
		}
	}

	return playersWithTag
}

// GetPlayersByTagCount returns players sorted by their count of a specific tag
func (tm *TagManager) GetPlayersByTagCount(ctx context.Context, players []model.Player, tag model.CardTag) []TagCountResult {
	var results []TagCountResult

	for _, player := range players {
		count := tm.GetPlayerTagCount(ctx, &player, tag)
		results = append(results, TagCountResult{
			Player: player,
			Count:  count,
		})
	}

	// Sort by count (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Count < results[j].Count {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// TagCountResult represents a player and their tag count
type TagCountResult struct {
	Player model.Player
	Count  int
}

// GetAllUniqueTagsInGame returns all unique tags present in a game across all players
func (tm *TagManager) GetAllUniqueTagsInGame(ctx context.Context, players []model.Player) map[model.CardTag]int {
	allTags := make(map[model.CardTag]int)

	for _, player := range players {
		playerTags := tm.CountPlayerTags(ctx, &player)
		for tag, count := range playerTags {
			allTags[tag] += count
		}
	}

	return allTags
}

// GetTagSynergies calculates potential tag synergies for a player
func (tm *TagManager) GetTagSynergies(ctx context.Context, player *model.Player) TagSynergyInfo {
	tagCounts := tm.CountPlayerTags(ctx, player)

	synergy := TagSynergyInfo{
		TagCounts:    tagCounts,
		TotalTags:    0,
		DiverseCount: len(tagCounts),
	}

	// Calculate total tag count
	for _, count := range tagCounts {
		synergy.TotalTags += count
	}

	// Find most common tag
	maxCount := 0
	for tag, count := range tagCounts {
		if count > maxCount {
			maxCount = count
			synergy.MostCommonTag = &tag
		}
	}
	synergy.MostCommonCount = maxCount

	return synergy
}

// TagSynergyInfo contains information about a player's tag synergies
type TagSynergyInfo struct {
	TagCounts       map[model.CardTag]int
	TotalTags       int
	DiverseCount    int
	MostCommonTag   *model.CardTag
	MostCommonCount int
}
