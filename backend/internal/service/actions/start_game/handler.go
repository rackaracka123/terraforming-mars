package start_game

import (
	"context"
	"fmt"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	
	"go.uber.org/zap"
)

// StartGameHandler handles start game actions
type StartGameHandler struct{
	EventBus events.EventBus
}

// Handle applies the start game action
func (h *StartGameHandler) Handle(game *model.Game, player *model.Player, actionRequest dto.ActionStartGameRequest) error {
	action := actionRequest.GetAction()
	return h.applyStartGame(game, player, *action)
}

// applyStartGame applies start game action
func (h *StartGameHandler) applyStartGame(game *model.Game, player *model.Player, action dto.StartGameAction) error {
	log := logger.WithGameContext(game.ID, player.ID)
	
	// Validate that the player is the host
	if !game.IsHost(player.ID) {
		return fmt.Errorf("only the host can start the game")
	}

	// Validate game can be started
	if game.Status != model.GameStatusLobby {
		return fmt.Errorf("game is not in lobby status")
	}

	if len(game.Players) < 1 {
		return fmt.Errorf("need at least 1 player to start the game")
	}

	// Start the game
	game.Status = model.GameStatusActive
	game.CurrentPhase = model.GamePhaseStartingCardSelection

	// Set first player as active
	if len(game.Players) > 0 {
		game.CurrentPlayerID = game.Players[0].ID
	}

	// Collect player IDs for the event
	playerIDs := make([]string, len(game.Players))
	for i, p := range game.Players {
		playerIDs[i] = p.ID
	}

	// Publish GameStarted event to trigger starting card selection
	gameStartedEvent := events.NewGameStartedEvent(game.ID, playerIDs)
	if err := h.EventBus.Publish(context.Background(), gameStartedEvent); err != nil {
		log.Error("Failed to publish game started event", zap.Error(err))
		return fmt.Errorf("failed to publish game started event: %w", err)
	}

	log.Info("Game started and event published", 
		zap.Int("player_count", len(game.Players)),
	)

	return nil
}