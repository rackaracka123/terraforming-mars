package common_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/usecase/common"

	"github.com/stretchr/testify/assert"
)

func TestActionValidator_ValidatePlayerAction_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	gameID := "test-game"
	playerID := "test-player"

	// Create repositories
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()

	// Create validator
	validator := common.NewActionValidator(gameRepo, playerRepo)

	// Set up test data
	settings := model.GameSettings{MaxPlayers: 2}
	game, _ := gameRepo.Create(ctx, settings)
	gameID = game.ID

	// Set game to active with current turn and remaining actions
	gameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive)
	gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
	gameRepo.UpdateRemainingActions(ctx, gameID, 1)
	gameRepo.AddPlayerID(ctx, gameID, playerID)

	// Create player with sufficient resources
	player := model.Player{
		ID:              playerID,
		Name:            "Test Player",
		TerraformRating: 20,
		Resources: model.Resources{
			Credits:  50,
			Steel:    10,
			Titanium: 5,
			Plants:   8,
			Energy:   6,
			Heat:     4,
		},
	}
	playerRepo.Create(ctx, gameID, player)

	cost := common.ActionCost{
		Credits:  14,
		Steel:    2,
		Titanium: 1,
		Plants:   3,
		Energy:   2,
		Heat:     1,
	}

	// Act
	err := validator.ValidatePlayerAction(ctx, gameID, playerID, cost)

	// Assert
	assert.NoError(t, err)
}

func TestActionValidator_ValidatePlayerAction_GameNotActive(t *testing.T) {
	// Arrange
	ctx := context.Background()
	gameID := "test-game"
	playerID := "test-player"

	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	validator := common.NewActionValidator(gameRepo, playerRepo)

	// Create game but keep it in lobby status (not active)
	settings := model.GameSettings{MaxPlayers: 2}
	game, _ := gameRepo.Create(ctx, settings)
	gameID = game.ID

	cost := common.ActionCost{Credits: 14}

	// Act
	err := validator.ValidatePlayerAction(ctx, gameID, playerID, cost)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "game is not active")
}

func TestActionValidator_ValidatePlayerAction_NotPlayerTurn(t *testing.T) {
	// Arrange
	ctx := context.Background()
	gameID := "test-game"
	playerID := "test-player"
	otherPlayerID := "other-player"

	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	validator := common.NewActionValidator(gameRepo, playerRepo)

	// Set up game with other player's turn
	settings := model.GameSettings{MaxPlayers: 2}
	game, _ := gameRepo.Create(ctx, settings)
	gameID = game.ID

	gameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive)
	gameRepo.SetCurrentTurn(ctx, gameID, &otherPlayerID) // Other player's turn
	gameRepo.UpdateRemainingActions(ctx, gameID, 1)

	cost := common.ActionCost{Credits: 14}

	// Act
	err := validator.ValidatePlayerAction(ctx, gameID, playerID, cost)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not player's turn")
}

func TestActionValidator_ValidatePlayerAction_NoRemainingActions(t *testing.T) {
	// Arrange
	ctx := context.Background()
	gameID := "test-game"
	playerID := "test-player"

	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	validator := common.NewActionValidator(gameRepo, playerRepo)

	// Set up game with no remaining actions
	settings := model.GameSettings{MaxPlayers: 2}
	game, _ := gameRepo.Create(ctx, settings)
	gameID = game.ID

	gameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive)
	gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
	gameRepo.UpdateRemainingActions(ctx, gameID, 0) // No remaining actions

	cost := common.ActionCost{Credits: 14}

	// Act
	err := validator.ValidatePlayerAction(ctx, gameID, playerID, cost)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no remaining actions")
}

func TestActionValidator_ValidatePlayerAction_InsufficientCredits(t *testing.T) {
	// Arrange
	ctx := context.Background()
	gameID := "test-game"
	playerID := "test-player"

	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	validator := common.NewActionValidator(gameRepo, playerRepo)

	// Set up valid game state
	settings := model.GameSettings{MaxPlayers: 2}
	game, _ := gameRepo.Create(ctx, settings)
	gameID = game.ID

	gameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive)
	gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
	gameRepo.UpdateRemainingActions(ctx, gameID, 1)
	gameRepo.AddPlayerID(ctx, gameID, playerID)

	// Create player with insufficient credits
	player := model.Player{
		ID:              playerID,
		Name:            "Test Player",
		TerraformRating: 20,
		Resources: model.Resources{
			Credits: 5, // Not enough for 14 credit cost
		},
	}
	playerRepo.Create(ctx, gameID, player)

	cost := common.ActionCost{Credits: 14}

	// Act
	err := validator.ValidatePlayerAction(ctx, gameID, playerID, cost)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient resources")
	assert.Contains(t, err.Error(), "need 14 credits, have 5")
}

func TestActionValidator_ValidatePlayerAction_InsufficientMultipleResources(t *testing.T) {
	// Arrange
	ctx := context.Background()
	gameID := "test-game"
	playerID := "test-player"

	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	validator := common.NewActionValidator(gameRepo, playerRepo)

	// Set up valid game state
	settings := model.GameSettings{MaxPlayers: 2}
	game, _ := gameRepo.Create(ctx, settings)
	gameID = game.ID

	gameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive)
	gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
	gameRepo.UpdateRemainingActions(ctx, gameID, 1)
	gameRepo.AddPlayerID(ctx, gameID, playerID)

	// Create player with insufficient steel
	player := model.Player{
		ID:              playerID,
		Name:            "Test Player",
		TerraformRating: 20,
		Resources: model.Resources{
			Credits:  50,
			Steel:    1, // Not enough steel
			Titanium: 5,
			Plants:   8,
			Energy:   6,
			Heat:     4,
		},
	}
	playerRepo.Create(ctx, gameID, player)

	cost := common.ActionCost{
		Credits:  14,
		Steel:    5, // Requires more steel than player has
		Titanium: 1,
	}

	// Act
	err := validator.ValidatePlayerAction(ctx, gameID, playerID, cost)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient resources")
	assert.Contains(t, err.Error(), "need 5 steel, have 1")
}
