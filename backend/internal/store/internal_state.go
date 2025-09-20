package store

import (
	"math/rand"
	"time"

	"terraforming-mars-backend/internal/model"
)

// Internal State Management Functions
// These functions handle business logic and internal state changes triggered by user actions.
// They are pure functions that take application state and return modified state.

// dealStartingCards deals starting cards to all players in a game
func dealStartingCards(state ApplicationState, gameID string, cardCount int) ApplicationState {
	if state.CardRegistry() == nil {
		return state
	}

	players := state.GetGamePlayers(gameID)
	newState := state
	for _, playerState := range players {
		startingCards := getRandomStartingCards(newState, cardCount)
		if startingCards != nil {
			player := playerState.Player()
			player.StartingSelection = startingCards
			updatedPlayerState := playerState.WithPlayer(player)
			newState = newState.WithPlayer(player.ID, updatedPlayerState)
		}
	}

	return newState
}

// advanceToNextPhase changes the game phase and updates relevant state
func advanceToNextPhase(state ApplicationState, gameID string, newPhase model.GamePhase) ApplicationState {
	gameState, exists := state.GetGame(gameID)
	if !exists {
		return state
	}

	game := gameState.Game()
	game.CurrentPhase = newPhase
	game.UpdatedAt = time.Now()

	updatedGameState := gameState.WithGame(game)
	return state.WithGame(gameID, updatedGameState)
}

// setCurrentTurn assigns the current turn to a specific player
func setCurrentTurn(state ApplicationState, gameID string, playerID *string) ApplicationState {
	gameState, exists := state.GetGame(gameID)
	if !exists {
		return state
	}

	game := gameState.Game()
	game.CurrentTurn = playerID
	game.UpdatedAt = time.Now()

	updatedGameState := gameState.WithGame(game)
	return state.WithGame(gameID, updatedGameState)
}

// executeProduction runs the production phase for all players and advances generation
func executeProduction(state ApplicationState, gameID string) ApplicationState {
	gameState, exists := state.GetGame(gameID)
	if !exists {
		return state
	}

	game := gameState.Game()

	// Execute production for all players
	players := state.GetGamePlayers(gameID)
	newState := state
	for _, playerState := range players {
		newState = executePlayerProduction(newState, playerState.Player())
	}

	// Advance generation and phase
	game.Generation++
	game.CurrentPhase = model.GamePhaseAction

	// Set first player turn for new generation
	if len(players) > 0 {
		firstPlayerID := players[0].Player().ID
		game.CurrentTurn = &firstPlayerID
	}

	game.UpdatedAt = time.Now()
	updatedGameState := gameState.WithGame(game)
	return newState.WithGame(gameID, updatedGameState)
}

// executePlayerProduction runs production for a single player
func executePlayerProduction(state ApplicationState, player model.Player) ApplicationState {
	// Convert energy to heat
	energyConverted := player.Resources.Energy

	// Calculate new resources from production
	newResources := model.Resources{
		Credits:  player.Resources.Credits + player.Production.Credits + player.TerraformRating,
		Steel:    player.Resources.Steel + player.Production.Steel,
		Titanium: player.Resources.Titanium + player.Production.Titanium,
		Plants:   player.Resources.Plants + player.Production.Plants,
		Energy:   player.Production.Energy, // Reset to production amount
		Heat:     player.Resources.Heat + energyConverted + player.Production.Heat,
	}

	updatedPlayer := player
	updatedPlayer.Resources = newResources
	updatedPlayer.Passed = false
	updatedPlayer.AvailableActions = 2 // Reset actions for new generation

	playerState, exists := state.GetPlayer(player.ID)
	if !exists {
		return state
	}

	updatedPlayerState := playerState.WithPlayer(updatedPlayer)
	return state.WithPlayer(player.ID, updatedPlayerState)
}

// updateGlobalParameters sets the game's global parameters
func updateGlobalParameters(state ApplicationState, gameID string, params model.GlobalParameters) ApplicationState {
	gameState, exists := state.GetGame(gameID)
	if !exists {
		return state
	}

	game := gameState.Game()
	game.GlobalParameters = params
	game.UpdatedAt = time.Now()

	updatedGameState := gameState.WithGame(game)
	return state.WithGame(gameID, updatedGameState)
}

// increaseTemperature increases the global temperature by specified steps
func increaseTemperature(state ApplicationState, gameID string, steps int) ApplicationState {
	gameState, exists := state.GetGame(gameID)
	if !exists {
		return state
	}

	game := gameState.Game()

	// Each step increases temperature by 2°C, max +8°C
	newTemp := game.GlobalParameters.Temperature + (steps * 2)
	if newTemp > 8 {
		newTemp = 8
	}

	game.GlobalParameters.Temperature = newTemp
	game.UpdatedAt = time.Now()

	updatedGameState := gameState.WithGame(game)
	return state.WithGame(gameID, updatedGameState)
}

// increaseOxygen increases the global oxygen by specified steps
func increaseOxygen(state ApplicationState, gameID string, steps int) ApplicationState {
	gameState, exists := state.GetGame(gameID)
	if !exists {
		return state
	}

	game := gameState.Game()

	// Each step increases oxygen by 1%, max 14%
	newOxygen := game.GlobalParameters.Oxygen + steps
	if newOxygen > 14 {
		newOxygen = 14
	}

	game.GlobalParameters.Oxygen = newOxygen
	game.UpdatedAt = time.Now()

	updatedGameState := gameState.WithGame(game)
	return state.WithGame(gameID, updatedGameState)
}

// placeOcean increases the ocean count by specified amount
func placeOcean(state ApplicationState, gameID string, count int) ApplicationState {
	gameState, exists := state.GetGame(gameID)
	if !exists {
		return state
	}

	game := gameState.Game()

	// Each ocean tile, max 9
	newOceans := game.GlobalParameters.Oceans + count
	if newOceans > 9 {
		newOceans = 9
	}

	game.GlobalParameters.Oceans = newOceans
	game.UpdatedAt = time.Now()

	updatedGameState := gameState.WithGame(game)
	return state.WithGame(gameID, updatedGameState)
}

// consumePlayerAction decreases a player's available actions
func consumePlayerAction(state ApplicationState, playerID string) ApplicationState {
	playerState, exists := state.GetPlayer(playerID)
	if !exists {
		return state
	}

	player := playerState.Player()
	if player.AvailableActions <= 0 {
		return state
	}

	player.AvailableActions--
	updatedPlayerState := playerState.WithPlayer(player)
	return state.WithPlayer(playerID, updatedPlayerState)
}

// resetPlayerActions resets a player's available actions to 2
func resetPlayerActions(state ApplicationState, playerID string) ApplicationState {
	playerState, exists := state.GetPlayer(playerID)
	if !exists {
		return state
	}

	player := playerState.Player()
	player.AvailableActions = 2
	updatedPlayerState := playerState.WithPlayer(player)
	return state.WithPlayer(playerID, updatedPlayerState)
}

// updatePlayerResources sets a player's resources
func updatePlayerResources(state ApplicationState, playerID string, resources model.Resources) ApplicationState {
	playerState, exists := state.GetPlayer(playerID)
	if !exists {
		return state
	}

	player := playerState.Player()
	player.Resources = resources
	updatedPlayerState := playerState.WithPlayer(player)
	return state.WithPlayer(playerID, updatedPlayerState)
}

// deductPlayerResources subtracts costs from a player's resources
func deductPlayerResources(state ApplicationState, playerID string, cost model.Resources) ApplicationState {
	playerState, exists := state.GetPlayer(playerID)
	if !exists {
		return state
	}

	player := playerState.Player()

	// Verify player has sufficient resources
	if !hasPlayerResources(player.Resources, cost) {
		return state // Don't modify if insufficient resources
	}

	player.Resources.Credits -= cost.Credits
	player.Resources.Steel -= cost.Steel
	player.Resources.Titanium -= cost.Titanium
	player.Resources.Plants -= cost.Plants
	player.Resources.Energy -= cost.Energy
	player.Resources.Heat -= cost.Heat

	updatedPlayerState := playerState.WithPlayer(player)
	return state.WithPlayer(playerID, updatedPlayerState)
}

// updatePlayerProduction sets a player's production
func updatePlayerProduction(state ApplicationState, playerID string, production model.Production) ApplicationState {
	playerState, exists := state.GetPlayer(playerID)
	if !exists {
		return state
	}

	player := playerState.Player()
	player.Production = production
	updatedPlayerState := playerState.WithPlayer(player)
	return state.WithPlayer(playerID, updatedPlayerState)
}

// updateTerraformRating sets a player's terraform rating
func updateTerraformRating(state ApplicationState, playerID string, rating int) ApplicationState {
	playerState, exists := state.GetPlayer(playerID)
	if !exists {
		return state
	}

	player := playerState.Player()
	player.TerraformRating = rating
	updatedPlayerState := playerState.WithPlayer(player)
	return state.WithPlayer(playerID, updatedPlayerState)
}

// advanceTurn finds the next player who can take actions and sets their turn
func advanceTurn(state ApplicationState, gameID string) ApplicationState {
	gameState, exists := state.GetGame(gameID)
	if !exists {
		return state
	}

	game := gameState.Game()
	players := state.GetGamePlayers(gameID)
	if len(players) == 0 {
		return state
	}

	// Check if all players have passed or have no actions left
	allFinished := true
	for _, playerState := range players {
		player := playerState.Player()
		if !player.Passed && player.AvailableActions > 0 {
			allFinished = false
			break
		}
	}

	if allFinished {
		// Start production phase
		newState := advanceToNextPhase(state, gameID, model.GamePhaseProductionAndCardDraw)
		newState = setCurrentTurn(newState, gameID, nil)
		return newState
	}

	// Find current player index
	currentIndex := -1
	if game.CurrentTurn != nil {
		for i, playerState := range players {
			if playerState.Player().ID == *game.CurrentTurn {
				currentIndex = i
				break
			}
		}
	}

	// Find next player with available actions
	nextIndex := (currentIndex + 1) % len(players)
	for i := 0; i < len(players); i++ {
		nextPlayerState := players[nextIndex]
		nextPlayer := nextPlayerState.Player()
		if !nextPlayer.Passed && nextPlayer.AvailableActions > 0 {
			return setCurrentTurn(state, gameID, &nextPlayer.ID)
		}
		nextIndex = (nextIndex + 1) % len(players)
	}

	return state
}

// hasPlayerResources checks if a player has sufficient resources for a cost
func hasPlayerResources(playerResources, cost model.Resources) bool {
	return playerResources.Credits >= cost.Credits &&
		playerResources.Steel >= cost.Steel &&
		playerResources.Titanium >= cost.Titanium &&
		playerResources.Plants >= cost.Plants &&
		playerResources.Energy >= cost.Energy &&
		playerResources.Heat >= cost.Heat
}

// getRandomStartingCards randomly selects a specified number of cards from the starting deck
func getRandomStartingCards(state ApplicationState, count int) []string {
	if state.CardRegistry() == nil {
		return nil
	}

	pool := state.CardRegistry().GetStartingCardPool()
	if len(pool) < count {
		return nil // Not enough cards in pool
	}

	// Create a slice of card IDs from the pool
	cardIDs := make([]string, len(pool))
	for i, card := range pool {
		cardIDs[i] = card.ID
	}

	// Shuffle the card IDs
	rand.Shuffle(len(cardIDs), func(i, j int) {
		cardIDs[i], cardIDs[j] = cardIDs[j], cardIDs[i]
	})

	// Return the first 'count' cards
	result := make([]string, count)
	copy(result, cardIDs[:count])
	return result
}
