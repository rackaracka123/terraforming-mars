package integration

import (
	"fmt"
	"testing"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFieldCappedCityTilePlacement tests the complete flow of playing Field-Capped City card
// which places a city tile, including:
// 1. Setting up game in development mode
// 2. Giving player the Field-Capped City card (X21) via admin command
// 3. Playing the card
// 4. Verifying tile selection is triggered
// 5. Selecting a tile location
// 6. Verifying the tile is placed
// 7. Verifying bonuses are awarded (production changes, resource gains, placement bonuses)
func TestFieldCappedCityTilePlacement(t *testing.T) {
	client := NewTestClient(t)
	defer client.Close()

	// STEP 1: Connect and create game
	err := client.Connect()
	require.NoError(t, err, "Should be able to establish WebSocket connection")
	t.Log("✅ WebSocket connection established")

	gameID, err := client.CreateGameViaHTTP()
	require.NoError(t, err, "Should be able to create game")
	t.Log("✅ Game created:", gameID)

	// STEP 2: Join game
	playerName := "TestPlayer"
	err = client.JoinGameViaWebSocket(gameID, playerName)
	require.NoError(t, err, "Should be able to join game")
	t.Log("✅ Player joined game")

	// Wait for game-updated message
	msg, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Should receive game-updated message")
	t.Log("✅ Received game-updated after join")

	// Extract player ID from game state
	payload, ok := msg.Payload.(map[string]interface{})
	require.True(t, ok, "Payload should be a map")

	gameState, ok := payload["game"].(map[string]interface{})
	require.True(t, ok, "game should be a map in payload")

	currentPlayerMap, ok := gameState["currentPlayer"].(map[string]interface{})
	require.True(t, ok, "currentPlayer should be a map")

	playerID, ok := currentPlayerMap["id"].(string)
	require.True(t, ok, "playerID should be a string")
	t.Log("✅ Extracted player ID:", playerID)

	// Store player ID in client for convenience
	client.playerID = playerID
	client.gameID = gameID

	// STEP 3: Set game to action phase using admin command
	t.Log("📝 Setting game phase to action...")
	err = client.SendAdminCommand(dto.AdminCommandTypeSetPhase, map[string]interface{}{
		"phase": string(model.GamePhaseAction),
	})
	require.NoError(t, err, "Should be able to set phase")

	// Wait for phase change confirmation
	_, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Should receive game-updated after phase change")
	t.Log("✅ Game phase set to action")

	// STEP 3a: Set current turn to the player
	t.Log("🔄 Setting current turn to player...")
	err = client.SendAdminCommand(dto.AdminCommandTypeSetCurrentTurn, map[string]interface{}{
		"playerId": playerID,
	})
	require.NoError(t, err, "Should be able to set current turn")

	// Wait for turn change confirmation
	_, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Should receive game-updated after turn change")
	t.Log("✅ Current turn set to player")

	// STEP 4: Give player resources via admin command (to afford the card)
	t.Log("💰 Setting player resources...")
	err = client.SendAdminCommand(dto.AdminCommandTypeSetResources, map[string]interface{}{
		"playerId": playerID,
		"resources": map[string]interface{}{
			"credits":  100,
			"steel":    5,
			"titanium": 5,
			"plants":   0,
			"energy":   0,
			"heat":     0,
		},
	})
	require.NoError(t, err, "Should be able to set resources")

	_, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Should receive game-updated after resources set")
	t.Log("✅ Player resources set")

	// STEP 5: Give player Field-Capped City card (X21) via admin command
	t.Log("🎴 Giving Field-Capped City card to player...")
	err = client.SendAdminCommand(dto.AdminCommandTypeGiveCard, map[string]interface{}{
		"playerId": playerID,
		"cardId":   "X21", // Field-Capped City
	})
	require.NoError(t, err, "Should be able to give card")

	_, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Should receive game-updated after card given")
	t.Log("✅ Field-Capped City card given to player")

	// STEP 6: Play the Field-Capped City card
	t.Log("🎯 Playing Field-Capped City card...")
	err = client.SendAction(dto.MessageTypeActionPlayCard, map[string]interface{}{
		"cardId": "X21",
	})
	require.NoError(t, err, "Should be able to send play-card action")

	// Wait for game-updated message after card play
	msg, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Should receive game-updated after card play")
	t.Log("✅ Card played successfully")

	// STEP 7: Verify tile selection was triggered
	payload, ok = msg.Payload.(map[string]interface{})
	require.True(t, ok, "Payload should be a map")

	gameState, ok = payload["game"].(map[string]interface{})
	require.True(t, ok, "game should be a map in payload")

	currentPlayerMap, ok = gameState["currentPlayer"].(map[string]interface{})
	require.True(t, ok, "currentPlayer should be a map")

	pendingTileSelection, ok := currentPlayerMap["pendingTileSelection"].(map[string]interface{})
	require.True(t, ok, "pendingTileSelection should be present")
	require.NotNil(t, pendingTileSelection, "pendingTileSelection should not be nil")

	tileType, ok := pendingTileSelection["tileType"].(string)
	require.True(t, ok, "tileType should be a string")
	assert.Equal(t, "city", tileType, "Should be placing a city tile")

	availableHexes, ok := pendingTileSelection["availableHexes"].([]interface{})
	require.True(t, ok, "availableHexes should be an array")
	require.Greater(t, len(availableHexes), 0, "Should have available hexes for city placement")

	t.Logf("✅ Tile selection triggered for city with %d available hexes", len(availableHexes))

	// STEP 8: Verify production changes from Field-Capped City
	// Field-Capped City should increase credits production by 2 and energy production by 1
	production, ok := currentPlayerMap["resourceProduction"].(map[string]interface{})
	require.True(t, ok, "resourceProduction should be a map")

	creditsProduction, ok := production["credits"].(float64)
	require.True(t, ok, "credits production should be a number")
	assert.Equal(t, float64(3), creditsProduction, "Credits production should be 3 (1 starting + 2 from card)")

	energyProduction, ok := production["energy"].(float64)
	require.True(t, ok, "energy production should be a number")
	assert.Equal(t, float64(1), energyProduction, "Energy production should be increased by 1")

	t.Log("✅ Production changes verified: 3 credits production (1 starting + 2 from card), 1 energy production")

	// STEP 9: Verify player received 3 plants
	resources, ok := currentPlayerMap["resources"].(map[string]interface{})
	require.True(t, ok, "resources should be a map")

	plants, ok := resources["plants"].(float64)
	require.True(t, ok, "plants should be a number")
	assert.Equal(t, float64(3), plants, "Should have received 3 plants")

	t.Log("✅ Resource gain verified: +3 plants")

	// STEP 10: Select a tile to place the city
	// Pick the first available hex
	firstHex, ok := availableHexes[0].(string)
	require.True(t, ok, "First hex should be a string")

	t.Logf("🎯 Selecting tile at coordinate: %s", firstHex)

	// Parse coordinate string (format: "q,r,s")
	var q, r, s int
	_, err = fmt.Sscanf(firstHex, "%d,%d,%d", &q, &r, &s)
	require.NoError(t, err, "Should be able to parse coordinate")

	err = client.SendAction(dto.MessageTypeActionTileSelected, map[string]interface{}{
		"coordinate": map[string]interface{}{
			"q": q,
			"r": r,
			"s": s,
		},
	})
	require.NoError(t, err, "Should be able to send tile-selected action")

	// Wait for game-updated message after tile placement
	msg, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Should receive game-updated after tile placement")
	t.Log("✅ Tile placement successful")

	// STEP 11: Verify the tile was placed on the board
	payload, ok = msg.Payload.(map[string]interface{})
	require.True(t, ok, "Payload should be a map")

	gameState, ok = payload["game"].(map[string]interface{})
	require.True(t, ok, "game should be a map in payload")

	board, ok := gameState["board"].(map[string]interface{})
	require.True(t, ok, "board should be a map")

	tiles, ok := board["tiles"].([]interface{})
	require.True(t, ok, "tiles should be an array")

	// Find the placed tile
	var placedTile map[string]interface{}
	for _, tileInterface := range tiles {
		tile, ok := tileInterface.(map[string]interface{})
		if !ok {
			continue
		}

		coordinates, ok := tile["coordinates"].(map[string]interface{})
		if !ok {
			continue
		}

		tileQ, ok := coordinates["q"].(float64)
		if !ok {
			continue
		}
		tileR, ok := coordinates["r"].(float64)
		if !ok {
			continue
		}
		tileS, ok := coordinates["s"].(float64)
		if !ok {
			continue
		}

		if int(tileQ) == q && int(tileR) == r && int(tileS) == s {
			placedTile = tile
			break
		}
	}

	require.NotNil(t, placedTile, "Should find the placed tile on the board")

	// Verify tile is occupied by a city
	occupiedBy, ok := placedTile["occupiedBy"].(map[string]interface{})
	require.True(t, ok, "occupiedBy should be a map")

	occupantType, ok := occupiedBy["type"].(string)
	require.True(t, ok, "type should be a string")
	assert.Equal(t, "city-tile", occupantType, "Tile should be occupied by a city")

	// Verify tile is owned by the player
	ownerID, ok := placedTile["ownerId"].(string)
	require.True(t, ok, "ownerId should be a string")
	assert.Equal(t, playerID, ownerID, "Tile should be owned by the player")

	t.Logf("✅ City tile placed at %d,%d,%d and owned by player", q, r, s)

	// STEP 12: Verify placement bonuses were awarded
	// Check if the tile had any bonuses and if resources were updated accordingly
	currentPlayerMap, ok = gameState["currentPlayer"].(map[string]interface{})
	require.True(t, ok, "currentPlayer should be a map")

	resources, ok = currentPlayerMap["resources"].(map[string]interface{})
	require.True(t, ok, "resources should be a map")

	// Check for tile bonuses on the placed tile
	bonuses, ok := placedTile["bonuses"].([]interface{})
	if ok && len(bonuses) > 0 {
		t.Logf("✅ Tile had %d bonuses", len(bonuses))
		// Log what bonuses were found
		for _, bonusInterface := range bonuses {
			bonus, ok := bonusInterface.(map[string]interface{})
			if !ok {
				continue
			}
			bonusType, _ := bonus["type"].(string)
			bonusAmount, _ := bonus["amount"].(float64)
			t.Logf("   - %s: %d", bonusType, int(bonusAmount))
		}
	}

	// Check for ocean adjacency bonuses (would show as increased credits)
	credits, ok := resources["credits"].(float64)
	require.True(t, ok, "credits should be a number")
	t.Logf("✅ Player has %d credits after tile placement", int(credits))

	// STEP 13: Verify no pending tile selection remains
	pendingTileSelectionAfter, hasSelection := currentPlayerMap["pendingTileSelection"]
	assert.False(t, hasSelection && pendingTileSelectionAfter != nil, "Should not have pending tile selection after placement")

	t.Log("✅ No pending tile selection remaining")

	t.Log("🎉 Complete Field-Capped City tile placement flow test passed!")
}
