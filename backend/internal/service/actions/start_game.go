package actions

import (
	"fmt"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/domain"
)

// StartGameHandler handles start game actions
type StartGameHandler struct{}

// Handle applies the start game action
func (h *StartGameHandler) Handle(game *domain.Game, player *domain.Player, actionPayload dto.ActionPayload) error {
	action := dto.StartGameAction{Type: actionPayload.Type}
	return h.applyStartGame(game, player, action)
}

// applyStartGame applies start game action
func (h *StartGameHandler) applyStartGame(game *domain.Game, player *domain.Player, action dto.StartGameAction) error {
	// Validate that the player is the host
	if !game.IsHost(player.ID) {
		return fmt.Errorf("only the host can start the game")
	}

	// Validate game can be started
	if game.Status != domain.GameStatusLobby {
		return fmt.Errorf("game is not in lobby status")
	}

	if len(game.Players) < 1 {
		return fmt.Errorf("need at least 1 player to start the game")
	}

	// Start the game
	game.Status = domain.GameStatusActive
	game.CurrentPhase = domain.GamePhaseStartingCardSelection

	// Set first player as active
	if len(game.Players) > 0 {
		game.CurrentPlayerID = game.Players[0].ID
	}

	return nil
}