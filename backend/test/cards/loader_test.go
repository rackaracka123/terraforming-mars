package cards

import (
	"testing"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardLoader_LoadCardsFromAssets(t *testing.T) {
	// Test loading cards from assets with proper path
	registry, err := cards.LoadCardsFromFile("../../assets/terraforming_mars_cards.json")
	require.NoError(t, err, "Should successfully load cards from assets")
	require.NotNil(t, registry, "Registry should not be nil")

	// Verify basic structure
	assert.Greater(t, len(registry.Cards), 0, "Should have cards loaded")
	assert.Greater(t, len(registry.Corporations), 0, "Should have corporations loaded")
	assert.Greater(t, len(registry.StartingDeck), 0, "Should have starting deck")

	// Verify some expected cards exist
	credicor, exists := registry.GetCard("B01")
	assert.True(t, exists, "CrediCor corporation should exist")
	assert.Equal(t, "CrediCor", credicor.Name)
	assert.Equal(t, "corporation", string(credicor.Type))
}

func TestStoreIntegration_CardRegistry(t *testing.T) {
	// Test that cards are loaded into the store correctly
	eventBus := events.NewInMemoryEventBus()

	appStore, err := store.InitializeStore(eventBus)
	require.NoError(t, err, "Store initialization should succeed")
	require.NotNil(t, appStore, "Store should not be nil")

	// Check that card registry was loaded
	state := appStore.GetState()
	require.NotNil(t, state.CardRegistry(), "Card registry should be loaded in state")

	// Verify cards are accessible
	card, exists := state.CardRegistry().GetCard("B01")
	assert.True(t, exists, "Should be able to get cards from registry")
	assert.Equal(t, "CrediCor", card.Name)
}

func TestCardRegistry_DeepCopy(t *testing.T) {
	// Load a registry with proper path
	registry, err := cards.LoadCardsFromFile("../../assets/terraforming_mars_cards.json")
	require.NoError(t, err, "Should load cards successfully")

	// Create deep copy
	copyRegistry := registry.DeepCopy()
	require.NotNil(t, copyRegistry, "Copy should not be nil")

	// Verify they're separate objects
	assert.NotSame(t, registry, copyRegistry, "Should be different objects")
	assert.NotSame(t, registry.Cards, copyRegistry.Cards, "Cards map should be different objects")
	assert.NotSame(t, registry.Corporations, copyRegistry.Corporations, "Corporations map should be different objects")
	assert.NotSame(t, registry.StartingDeck, copyRegistry.StartingDeck, "Starting deck should be different objects")

	// Verify content is the same
	assert.Equal(t, len(registry.Cards), len(copyRegistry.Cards))
	assert.Equal(t, len(registry.Corporations), len(copyRegistry.Corporations))
	assert.Equal(t, len(registry.StartingDeck), len(copyRegistry.StartingDeck))
}
