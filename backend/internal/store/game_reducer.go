package store

import (
	"fmt"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
)

// GameReducer handles all game-related state changes
func GameReducer(state ApplicationState, action Action) (ApplicationState, error) {

	switch action.Type {
	case ActionCreateGame:
		return handleCreateGame(state, action)

	case ActionStartGame:
		return handleStartGame(state, action)

	case ActionJoinGame:
		return handleJoinGame(state, action)

	case ActionLeaveGame:
		return handleLeaveGame(state, action)

	case ActionSetPassed:
		return handleSetPassed(state, action)

	case ActionPlayCard:
		return handlePlayCard(state, action)

	case ActionSelectStartingCards:
		return handleSelectStartingCards(state, action)

	// Standard Project Actions
	case ActionBuildPowerPlant, ActionLaunchAsteroid, ActionBuildAquifer,
		ActionPlantGreenery, ActionBuildCity, ActionSellPatents:
		return handleStandardProject(state, action)

	default:
		// Not a game action, return unchanged state
		return state, nil
	}
}

func handleCreateGame(state ApplicationState, action Action) (ApplicationState, error) {
	payload, ok := action.Payload.(CreateGamePayload)
	if !ok {
		return state, fmt.Errorf("invalid payload for CREATE_GAME")
	}

	// Check if game already exists
	if _, exists := state.GetGame(payload.GameID); exists {
		return state, fmt.Errorf("game %s already exists", payload.GameID)
	}

	now := time.Now()
	game := model.Game{
		ID:           payload.GameID,
		CreatedAt:    now,
		UpdatedAt:    now,
		Status:       model.GameStatusLobby,
		Settings:     payload.Settings,
		PlayerIDs:    make([]string, 0),
		CurrentPhase: model.GamePhaseWaitingForGameStart,
		GlobalParameters: model.GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		},
		Generation: 1,
	}

	gameState := NewGameState(game)
	return state.WithGame(payload.GameID, gameState), nil
}

func handleStartGame(state ApplicationState, action Action) (ApplicationState, error) {
	// Extract gameID and playerID from action meta since DTO doesn't contain them
	gameID := action.Meta.GameID
	playerID := action.Meta.PlayerID

	if gameID == "" {
		return state, fmt.Errorf("gameID required for START_GAME")
	}

	gameState, exists := state.GetGame(gameID)
	if !exists {
		return state, fmt.Errorf("game %s not found", gameID)
	}

	game := gameState.Game()
	if game.HostPlayerID != playerID {
		return state, fmt.Errorf("only host can start game")
	}

	if game.Status != model.GameStatusLobby {
		return state, fmt.Errorf("game is not in lobby state")
	}

	players := state.GetGamePlayers(gameID)
	if len(players) < 1 {
		return state, fmt.Errorf("cannot start game with no players")
	}

	// Update game state
	updatedGame := game
	updatedGame.Status = model.GameStatusActive
	updatedGame.CurrentPhase = model.GamePhaseStartingCardSelection
	updatedGame.UpdatedAt = time.Now()

	// Set first player turn
	if len(players) > 0 {
		firstPlayerID := players[0].Player().ID
		updatedGame.CurrentTurn = &firstPlayerID
	}

	updatedGameState := gameState.WithGame(updatedGame)
	newState := state.WithGame(gameID, updatedGameState)

	// Deal starting cards to all players using business logic layer
	newState = dealStartingCards(newState, gameID, 10)

	return newState, nil
}

// Player Handler Functions (merged from PlayerReducer)

func handleJoinGame(state ApplicationState, action Action) (ApplicationState, error) {
	payload, ok := action.Payload.(JoinGamePayload)
	if !ok {
		return state, fmt.Errorf("invalid payload for JOIN_GAME")
	}

	gameState, exists := state.GetGame(payload.GameID)
	if !exists {
		return state, fmt.Errorf("game %s not found", payload.GameID)
	}

	game := gameState.Game()

	// Check if game is joinable
	if game.Status == model.GameStatusCompleted {
		return state, fmt.Errorf("cannot join completed game")
	}

	// Check if game is full
	if len(game.PlayerIDs) >= game.Settings.MaxPlayers {
		return state, fmt.Errorf("game is full")
	}

	// Check if player already exists with the same ID
	if existingPlayerState, exists := state.GetPlayer(payload.PlayerID); exists {
		// Player with same ID already exists - reactivate them for reconnection
		existingPlayer := existingPlayerState.Player()
		existingPlayer.IsConnected = true
		existingPlayer.Name = payload.PlayerName // Update name in case it changed

		updatedPlayerState := existingPlayerState.WithPlayer(existingPlayer)
		newState := state.WithPlayer(payload.PlayerID, updatedPlayerState)

		// Update game timestamp
		updatedGame := game
		updatedGame.UpdatedAt = time.Now()
		updatedGameState := gameState.WithGame(updatedGame)

		return newState.WithGame(payload.GameID, updatedGameState), nil
	}

	// Check if player with same name already exists in this game (potential reconnection with different ID)
	for _, playerID := range game.PlayerIDs {
		if existingPlayerState, exists := state.GetPlayer(playerID); exists {
			existingPlayer := existingPlayerState.Player()
			if existingPlayer.Name == payload.PlayerName {
				// Same name found - this is likely a reconnection attempt, don't create duplicate
				existingPlayer.IsConnected = true

				updatedPlayerState := existingPlayerState.WithPlayer(existingPlayer)
				newState := state.WithPlayer(playerID, updatedPlayerState)

				// Update game timestamp
				updatedGame := game
				updatedGame.UpdatedAt = time.Now()
				updatedGameState := gameState.WithGame(updatedGame)

				return newState.WithGame(payload.GameID, updatedGameState), nil
			}
		}
	}

	// Create new player
	player := model.Player{
		ID:   payload.PlayerID,
		Name: payload.PlayerName,
		Resources: model.Resources{
			Credits: 40, // Starting credits
		},
		Production: model.Production{
			Credits: 1, // Base production
		},
		TerraformRating:  20, // Starting terraform rating
		PlayedCards:      make([]string, 0),
		Passed:           false,
		AvailableActions: 2, // Standard actions per turn
		VictoryPoints:    0,
		IsConnected:      true,
	}

	playerState := NewPlayerState(player, payload.GameID)
	newState := state.WithPlayer(payload.PlayerID, playerState)

	// Add player to game's PlayerIDs list and update game
	updatedGame := game
	updatedGame.PlayerIDs = append(updatedGame.PlayerIDs, payload.PlayerID)

	// Set host if first player
	if updatedGame.HostPlayerID == "" {
		updatedGame.HostPlayerID = payload.PlayerID
	}

	updatedGameState := gameState.WithGame(updatedGame).WithPlayer(payload.PlayerID)
	return newState.WithGame(payload.GameID, updatedGameState), nil
}

func handleLeaveGame(state ApplicationState, action Action) (ApplicationState, error) {
	playerID := action.Meta.PlayerID
	return state.WithoutPlayer(playerID), nil
}

func handleSetPassed(state ApplicationState, action Action) (ApplicationState, error) {
	playerID := action.Meta.PlayerID
	if playerID == "" {
		return state, fmt.Errorf("playerID required for SET_PASSED")
	}

	playerState, exists := state.GetPlayer(playerID)
	if !exists {
		return state, fmt.Errorf("player %s not found", playerID)
	}

	// Extract passed value from payload - payload should be a boolean
	passed, ok := action.Payload.(bool)
	if !ok {
		return state, fmt.Errorf("invalid payload for SET_PASSED - expected boolean")
	}

	player := playerState.Player()
	player.Passed = passed
	updatedPlayerState := playerState.WithPlayer(player)
	return state.WithPlayer(playerID, updatedPlayerState), nil
}

func handlePlayCard(state ApplicationState, action Action) (ApplicationState, error) {
	payload, ok := action.Payload.(dto.PlayCardAction)
	if !ok {
		return state, fmt.Errorf("invalid payload for PLAY_CARD")
	}

	playerID := action.Meta.PlayerID
	if playerID == "" {
		return state, fmt.Errorf("playerID required for PLAY_CARD")
	}

	playerState, exists := state.GetPlayer(playerID)
	if !exists {
		return state, fmt.Errorf("player %s not found", playerID)
	}

	player := playerState.Player()
	player.PlayedCards = append(player.PlayedCards, payload.CardID)
	updatedPlayerState := playerState.WithPlayer(player)
	return state.WithPlayer(playerID, updatedPlayerState), nil
}

func handleSelectStartingCards(state ApplicationState, action Action) (ApplicationState, error) {
	payload, ok := action.Payload.(SelectStartingCardsPayload)
	if !ok {
		return state, fmt.Errorf("invalid payload for SELECT_STARTING_CARDS")
	}

	playerState, exists := state.GetPlayer(payload.PlayerID)
	if !exists {
		return state, fmt.Errorf("player %s not found", payload.PlayerID)
	}

	// Deduct cost for selected cards using business logic
	newState := state
	if payload.Cost > 0 {
		cost := model.Resources{Credits: payload.Cost}
		newState = deductPlayerResources(newState, payload.PlayerID, cost)
	}

	// Get player reference again after potential state update
	playerState, exists = newState.GetPlayer(payload.PlayerID)
	if !exists {
		return newState, fmt.Errorf("player %s not found after resource deduction", payload.PlayerID)
	}

	// Create updated player with cards added and starting selection cleared
	player := playerState.Player()
	player.Cards = append(player.Cards, payload.SelectedCards...)
	player.StartingSelection = nil

	updatedPlayerState := playerState.WithPlayer(player)
	newState = newState.WithPlayer(payload.PlayerID, updatedPlayerState)

	// Check if all players have completed their starting card selection
	gameState, exists := newState.GetGame(payload.GameID)
	if !exists {
		return newState, fmt.Errorf("game %s not found", payload.GameID)
	}

	game := gameState.Game()

	// Check if all players have completed starting card selection (no remaining startingSelection)
	allPlayersReady := true
	for _, playerID := range game.PlayerIDs {
		if playerState, exists := newState.GetPlayer(playerID); exists {
			if playerState.Player().StartingSelection != nil && len(playerState.Player().StartingSelection) > 0 {
				allPlayersReady = false
				break
			}
		}
	}

	// If all players have completed their starting card selection, transition to action phase
	if allPlayersReady {
		updatedGame := game
		updatedGame.CurrentPhase = model.GamePhaseAction
		updatedGame.UpdatedAt = time.Now()

		// Ensure current turn is set to first player if not already set
		if updatedGame.CurrentTurn == nil && len(updatedGame.PlayerIDs) > 0 {
			firstPlayerID := updatedGame.PlayerIDs[0]
			updatedGame.CurrentTurn = &firstPlayerID
		}

		// Set available actions for all players (2 for current turn player, 0 for others)
		for _, playerID := range updatedGame.PlayerIDs {
			if playerState, exists := newState.GetPlayer(playerID); exists {
				player := playerState.Player()
				if updatedGame.CurrentTurn != nil && playerID == *updatedGame.CurrentTurn {
					player.AvailableActions = 2 // Standard 2 actions per turn in Terraforming Mars
				} else {
					player.AvailableActions = 0 // Not their turn
				}
				updatedPlayerState := playerState.WithPlayer(player)
				newState = newState.WithPlayer(playerID, updatedPlayerState)
			}
		}

		updatedGameState := gameState.WithGame(updatedGame)
		newState = newState.WithGame(payload.GameID, updatedGameState)
	}

	return newState, nil
}

func handleStandardProject(state ApplicationState, action Action) (ApplicationState, error) {
	playerID := action.Meta.PlayerID
	if playerID == "" {
		return state, fmt.Errorf("playerID required for standard project")
	}

	playerState, exists := state.GetPlayer(playerID)
	if !exists {
		return state, fmt.Errorf("player %s not found", playerID)
	}

	player := playerState.Player()

	// Determine cost based on action type
	var cost model.Resources
	switch action.Type {
	case ActionBuildPowerPlant:
		cost = model.Resources{Credits: 11}
	case ActionSellPatents:
		cost = model.Resources{}
	case ActionLaunchAsteroid:
		cost = model.Resources{Credits: 14}
	case ActionBuildAquifer:
		cost = model.Resources{Credits: 18}
	case ActionPlantGreenery:
		cost = model.Resources{Credits: 23}
	case ActionBuildCity:
		cost = model.Resources{Credits: 25}
	default:
		return state, fmt.Errorf("unknown standard project type: %s", action.Type)
	}

	// Validate player has sufficient resources
	if !hasPlayerResources(player.Resources, cost) {
		return state, fmt.Errorf("insufficient resources for standard project")
	}

	// Deduct resources using business logic layer
	newState := deductPlayerResources(state, playerID, cost)

	// Apply standard project effects using business logic
	switch action.Type {
	case ActionBuildPowerPlant:
		newProduction := player.Production
		newProduction.Energy++
		newState = updatePlayerProduction(newState, playerID, newProduction)

	case ActionSellPatents:
		// No additional effects for sell patents

	case ActionLaunchAsteroid:
		// No additional effects for launch asteroid

	case ActionBuildAquifer:
		// No additional effects for build aquifer

	case ActionPlantGreenery:
		newProduction := player.Production
		newProduction.Plants++
		newState = updatePlayerProduction(newState, playerID, newProduction)

	case ActionBuildCity:
		newProduction := player.Production
		newProduction.Credits++
		newState = updatePlayerProduction(newState, playerID, newProduction)
	}

	return newState, nil
}
