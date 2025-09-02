package usecase

import (
	"fmt"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/repository"
	"time"
)

// GameUseCase handles game business logic
type GameUseCase struct {
	gameRepo *repository.GameRepository
}

// NewGameUseCase creates a new game use case
func NewGameUseCase(gameRepo *repository.GameRepository) *GameUseCase {
	return &GameUseCase{
		gameRepo: gameRepo,
	}
}

// GetGame retrieves a game by ID
func (uc *GameUseCase) GetGame(gameID string) (*domain.GameState, error) {
	return uc.gameRepo.GetGame(gameID)
}

// JoinGame adds a player to a game
func (uc *GameUseCase) JoinGame(gameID, playerID, playerName string) (*domain.GameState, error) {
	game, err := uc.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, err
	}
	
	// Check if player already exists
	for _, player := range game.Players {
		if player.ID == playerID {
			return game, nil // Player already in game
		}
	}
	
	// Create new player
	newPlayer := domain.Player{
		ID:   playerID,
		Name: playerName,
		Resources: domain.ResourcesMap{
			Credits:  20,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		Production: domain.ResourcesMap{
			Credits:  1,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   1,
			Heat:     1,
		},
		TerraformRating:   20,
		VictoryPoints:     0,
		PlayedCards:       []string{},
		Hand:              []string{},
		AvailableActions:  2,
		Tags:              []domain.Tag{},
		ActionsTaken:      0,
		ActionsRemaining:  2,
		TilePositions:     []domain.HexCoordinate{},
		Reserved:          domain.ResourcesMap{},
	}
	
	game.Players = append(game.Players, newPlayer)
	
	// Set current player if none set
	if game.CurrentPlayer == "" {
		game.CurrentPlayer = playerID
	}
	
	game.UpdatedAt = time.Now()
	
	return game, uc.gameRepo.SaveGame(game)
}

// SelectCorporation assigns a corporation to a player
func (uc *GameUseCase) SelectCorporation(gameID, playerID, corporationID string) (*domain.GameState, error) {
	game, err := uc.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, err
	}
	
	// Find the player
	var player *domain.Player
	for i := range game.Players {
		if game.Players[i].ID == playerID {
			player = &game.Players[i]
			break
		}
	}
	
	if player == nil {
		return nil, fmt.Errorf("player %s not found in game %s", playerID, gameID)
	}
	
	// Get corporation data
	corporations := domain.GetBaseCorporations()
	var selectedCorp *domain.Corporation
	for _, corp := range corporations {
		if corp.ID == corporationID {
			selectedCorp = &corp
			break
		}
	}
	
	if selectedCorp == nil {
		return nil, fmt.Errorf("corporation %s not found", corporationID)
	}
	
	// Apply corporation effects
	player.Corporation = &corporationID
	player.Resources.Credits = selectedCorp.StartingMegaCredits
	
	// Apply starting production
	if selectedCorp.StartingProduction != nil {
		player.Production.Credits += selectedCorp.StartingProduction.Credits
		player.Production.Steel += selectedCorp.StartingProduction.Steel
		player.Production.Titanium += selectedCorp.StartingProduction.Titanium
		player.Production.Plants += selectedCorp.StartingProduction.Plants
		player.Production.Energy += selectedCorp.StartingProduction.Energy
		player.Production.Heat += selectedCorp.StartingProduction.Heat
	}
	
	// Apply starting resources
	if selectedCorp.StartingResources != nil {
		player.Resources.Credits += selectedCorp.StartingResources.Credits
		player.Resources.Steel += selectedCorp.StartingResources.Steel
		player.Resources.Titanium += selectedCorp.StartingResources.Titanium
		player.Resources.Plants += selectedCorp.StartingResources.Plants
		player.Resources.Energy += selectedCorp.StartingResources.Energy
		player.Resources.Heat += selectedCorp.StartingResources.Heat
	}
	
	// Apply starting TR bonus
	if selectedCorp.StartingTR != nil {
		player.TerraformRating += *selectedCorp.StartingTR
	}
	
	game.UpdatedAt = time.Now()
	
	return game, uc.gameRepo.SaveGame(game)
}

// RaiseTemperature increases global temperature using heat
func (uc *GameUseCase) RaiseTemperature(gameID, playerID string) (*domain.GameState, error) {
	game, err := uc.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, err
	}
	
	// Find the player
	var player *domain.Player
	for i := range game.Players {
		if game.Players[i].ID == playerID {
			player = &game.Players[i]
			break
		}
	}
	
	if player == nil {
		return nil, fmt.Errorf("player %s not found in game %s", playerID, gameID)
	}
	
	// Check if it's the player's turn
	if game.CurrentPlayer != playerID {
		return nil, fmt.Errorf("not player's turn")
	}
	
	// Check if player has enough heat and actions
	if player.Resources.Heat < 8 {
		return nil, fmt.Errorf("not enough heat (need 8, have %d)", player.Resources.Heat)
	}
	
	if player.ActionsRemaining <= 0 {
		return nil, fmt.Errorf("no actions remaining")
	}
	
	// Check if temperature can be raised
	if game.GlobalParameters.Temperature >= 8 {
		return nil, fmt.Errorf("temperature already at maximum")
	}
	
	// Perform the action
	player.Resources.Heat -= 8
	game.GlobalParameters.Temperature += 2
	player.TerraformRating += 1
	
	// Update action tracking
	player.ActionsTaken += 1
	player.ActionsRemaining -= 1
	if game.CurrentActionCount != nil {
		*game.CurrentActionCount += 1
	}
	
	game.UpdatedAt = time.Now()
	
	return game, uc.gameRepo.SaveGame(game)
}

// SkipAction passes the turn to the next player
func (uc *GameUseCase) SkipAction(gameID, playerID string) (*domain.GameState, error) {
	game, err := uc.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, err
	}
	
	// Find the player
	var player *domain.Player
	for i := range game.Players {
		if game.Players[i].ID == playerID {
			player = &game.Players[i]
			break
		}
	}
	
	if player == nil {
		return nil, fmt.Errorf("player %s not found in game %s", playerID, gameID)
	}
	
	// Check if it's the player's turn
	if game.CurrentPlayer != playerID {
		return nil, fmt.Errorf("not player's turn")
	}
	
	// Mark player as passed
	player.Passed = boolPtr(true)
	player.ActionsRemaining = 0
	
	// Move to next player
	currentIndex := -1
	for i, p := range game.Players {
		if p.ID == game.CurrentPlayer {
			currentIndex = i
			break
		}
	}
	
	if currentIndex == -1 {
		return nil, fmt.Errorf("current player not found")
	}
	
	nextIndex := (currentIndex + 1) % len(game.Players)
	game.CurrentPlayer = game.Players[nextIndex].ID
	
	// Reset action tracking for next player
	nextPlayer := &game.Players[nextIndex]
	nextPlayer.ActionsTaken = 0
	nextPlayer.ActionsRemaining = 2
	
	if game.CurrentActionCount != nil {
		*game.CurrentActionCount = 0
	}
	
	game.UpdatedAt = time.Now()
	
	return game, uc.gameRepo.SaveGame(game)
}

// GetAvailableCorporations returns the list of available corporations
func (uc *GameUseCase) GetAvailableCorporations() []domain.Corporation {
	corporations := domain.GetBaseCorporations()
	
	// Add logo paths
	for i := range corporations {
		corporations[i].LogoPath = "/assets/misc/corpCard.png" // Generic corp card for now
	}
	
	return corporations
}

// boolPtr returns a pointer to a boolean
func boolPtr(b bool) *bool {
	return &b
}