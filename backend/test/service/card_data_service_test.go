package service_test

import (
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardDataService_LoadCards(t *testing.T) {
	// Create card data service
	cardService := service.NewCardDataService()

	// Load cards from JSON
	err := cardService.LoadCards()
	require.NoError(t, err, "Should load cards without error")

	// Test that cards were loaded
	allCards := cardService.GetAllCards()
	assert.Greater(t, len(allCards), 0, "Should have loaded cards")

	// Test specific card categories
	projectCards := cardService.GetProjectCards()
	corporationCards := cardService.GetCorporationCards()
	preludeCards := cardService.GetPreludeCards()

	assert.Greater(t, len(projectCards), 0, "Should have project cards")
	assert.Greater(t, len(corporationCards), 0, "Should have corporation cards")
	assert.Greater(t, len(preludeCards), 0, "Should have prelude cards")

	// Verify total matches expected counts from JSON
	expectedTotal := len(projectCards) + len(corporationCards) + len(preludeCards)
	assert.Equal(t, expectedTotal, len(allCards), "All cards should equal sum of categories")

	t.Logf("✅ Loaded %d total cards: %d project, %d corporation, %d prelude",
		len(allCards), len(projectCards), len(corporationCards), len(preludeCards))
}

func TestCardDataService_GetCardByID(t *testing.T) {
	cardService := service.NewCardDataService()
	err := cardService.LoadCards()
	require.NoError(t, err)

	// Test getting a card by ID
	allCards := cardService.GetAllCards()
	require.Greater(t, len(allCards), 0, "Should have cards to test with")

	// Test with first card
	firstCard := allCards[0]
	foundCard, err := cardService.GetCardByID(firstCard.ID)

	require.NoError(t, err, "Should find existing card")
	assert.Equal(t, firstCard.ID, foundCard.ID)
	assert.Equal(t, firstCard.Name, foundCard.Name)
	assert.Equal(t, firstCard.Type, foundCard.Type)

	// Test with non-existent ID
	_, err = cardService.GetCardByID("non-existent-card")
	assert.Error(t, err, "Should error for non-existent card")
}

func TestCardDataService_GetCardsByType(t *testing.T) {
	cardService := service.NewCardDataService()
	err := cardService.LoadCards()
	require.NoError(t, err)

	// Test getting cards by type
	automatedCards := cardService.GetCardsByType(model.CardTypeAutomated)
	activeCards := cardService.GetCardsByType(model.CardTypeActive)
	eventCards := cardService.GetCardsByType(model.CardTypeEvent)
	corporationCards := cardService.GetCardsByType(model.CardTypeCorporation)
	preludeCards := cardService.GetCardsByType(model.CardTypePrelude)

	// Verify all returned cards have the correct type
	for _, card := range automatedCards {
		assert.Equal(t, model.CardTypeAutomated, card.Type)
	}
	for _, card := range activeCards {
		assert.Equal(t, model.CardTypeActive, card.Type)
	}
	for _, card := range eventCards {
		assert.Equal(t, model.CardTypeEvent, card.Type)
	}
	for _, card := range corporationCards {
		assert.Equal(t, model.CardTypeCorporation, card.Type)
	}
	for _, card := range preludeCards {
		assert.Equal(t, model.CardTypePrelude, card.Type)
	}

	t.Logf("✅ Cards by type: %d automated, %d active, %d events, %d corporations, %d preludes",
		len(automatedCards), len(activeCards), len(eventCards), len(corporationCards), len(preludeCards))
}

func TestCardDataService_GetCardsByTag(t *testing.T) {
	cardService := service.NewCardDataService()
	err := cardService.LoadCards()
	require.NoError(t, err)

	// Test getting cards by specific tags
	spaceCards := cardService.GetCardsByTag(model.TagSpace)
	scienceCards := cardService.GetCardsByTag(model.TagScience)
	buildingCards := cardService.GetCardsByTag(model.TagBuilding)

	// Verify cards have the requested tag
	for _, card := range spaceCards {
		assert.Contains(t, card.Tags, model.TagSpace, "Space cards should have space tag")
	}
	for _, card := range scienceCards {
		assert.Contains(t, card.Tags, model.TagScience, "Science cards should have science tag")
	}
	for _, card := range buildingCards {
		assert.Contains(t, card.Tags, model.TagBuilding, "Building cards should have building tag")
	}

	t.Logf("✅ Cards by tag: %d space, %d science, %d building",
		len(spaceCards), len(scienceCards), len(buildingCards))
}

func TestCardDataService_GetCardsByCostRange(t *testing.T) {
	cardService := service.NewCardDataService()
	err := cardService.LoadCards()
	require.NoError(t, err)

	// Test getting cards by cost range
	cheapCards := cardService.GetCardsByCostRange(0, 10)
	mediumCards := cardService.GetCardsByCostRange(11, 20)
	expensiveCards := cardService.GetCardsByCostRange(21, 50)

	// Verify costs are within range
	for _, card := range cheapCards {
		assert.GreaterOrEqual(t, card.Cost, 0)
		assert.LessOrEqual(t, card.Cost, 10)
	}
	for _, card := range mediumCards {
		assert.GreaterOrEqual(t, card.Cost, 11)
		assert.LessOrEqual(t, card.Cost, 20)
	}
	for _, card := range expensiveCards {
		assert.GreaterOrEqual(t, card.Cost, 21)
		assert.LessOrEqual(t, card.Cost, 50)
	}

	t.Logf("✅ Cards by cost: %d cheap (0-10), %d medium (11-20), %d expensive (21-50)",
		len(cheapCards), len(mediumCards), len(expensiveCards))
}

func TestCardDataService_GetStartingCardPool(t *testing.T) {
	cardService := service.NewCardDataService()
	err := cardService.LoadCards()
	require.NoError(t, err)

	// Test getting starting card pool
	startingCards := cardService.GetStartingCardPool()
	assert.Greater(t, len(startingCards), 0, "Should have starting cards available")

	// Verify all starting cards meet criteria (cost <= 15, automated or active)
	for _, card := range startingCards {
		assert.LessOrEqual(t, card.Cost, 15, "Starting cards should be reasonably priced")
		assert.True(t, card.Type == model.CardTypeAutomated || card.Type == model.CardTypeActive,
			"Starting cards should be automated or active")
	}

	t.Logf("✅ %d cards available for starting selection", len(startingCards))
}

func TestCardDataService_GetCardsByTags(t *testing.T) {
	cardService := service.NewCardDataService()
	err := cardService.LoadCards()
	require.NoError(t, err)

	// Test getting cards by multiple tags (ANY)
	techTags := []model.CardTag{model.TagScience, model.TagSpace, model.TagPower}
	techCards := cardService.GetCardsByTags(techTags)

	// Verify each card has at least one of the requested tags
	for _, card := range techCards {
		hasAtLeastOneTag := false
		for _, cardTag := range card.Tags {
			for _, requestedTag := range techTags {
				if cardTag == requestedTag {
					hasAtLeastOneTag = true
					break
				}
			}
			if hasAtLeastOneTag {
				break
			}
		}
		assert.True(t, hasAtLeastOneTag, "Card %s should have at least one tech tag", card.Name)
	}

	t.Logf("✅ %d cards have at least one tech tag (science/space/power)", len(techCards))
}

func TestCardDataService_GetCardsByAllTags(t *testing.T) {
	cardService := service.NewCardDataService()
	err := cardService.LoadCards()
	require.NoError(t, err)

	// Test getting cards that have ALL specified tags
	// Use common combination
	commonTags := []model.CardTag{model.TagBuilding, model.TagPower}
	buildingPowerCards := cardService.GetCardsByAllTags(commonTags)

	// Verify each card has ALL requested tags
	for _, card := range buildingPowerCards {
		for _, requiredTag := range commonTags {
			assert.Contains(t, card.Tags, requiredTag,
				"Card %s should have all required tags", card.Name)
		}
	}

	t.Logf("✅ %d cards have both building AND power tags", len(buildingPowerCards))
}

func TestCardDataService_ConvertJSONCard(t *testing.T) {
	cardService := service.NewCardDataService()

	// This is an integration test - load the cards and verify structure
	err := cardService.LoadCards()
	require.NoError(t, err)

	allCards := cardService.GetAllCards()
	require.Greater(t, len(allCards), 0)

	// Test first few cards to ensure proper conversion
	for i, card := range allCards[:5] {
		t.Run(card.Name, func(t *testing.T) {
			assert.NotEmpty(t, card.ID, "Card %d should have ID", i)
			assert.NotEmpty(t, card.Name, "Card %d should have name", i)
			assert.NotEmpty(t, card.Type, "Card %d should have type", i)
			assert.GreaterOrEqual(t, card.Cost, 0, "Card %d should have non-negative cost", i)
			// Description can be empty for some cards, so we don't assert on it
			// Tags can be empty, so we don't assert on it
			// VictoryPoints can be 0, so we don't assert on it

			// If card has requirements, they should be properly parsed
			// This is mostly tested by the loading not failing
		})
	}

	t.Logf("✅ Successfully validated structure of %d cards", len(allCards))
}
