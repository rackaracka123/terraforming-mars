package start_game

import (
	"context"

	"terraforming-mars-backend/internal/service"
)

// Handler handles start game action requests
type Handler struct {
	gameService service.GameService
}

// NewHandler creates a new start game handler
func NewHandler(gameService service.GameService) *Handler {
	return &Handler{
		gameService: gameService,
	}
}

// Handle processes the start game action
func (h *Handler) Handle(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	return h.gameService.StartGame(ctx, gameID, playerID)
}
