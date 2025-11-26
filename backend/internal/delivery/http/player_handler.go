package http

import (
	"net/http"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/action/query"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/session"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// PlayerHandler handles HTTP requests related to player operations
type PlayerHandler struct {
	*BaseHandler
	sessionFactory  session.SessionFactory
	joinGameAction  *action.JoinGameAction
	getPlayerAction *query.GetPlayerAction
}

// NewPlayerHandler creates a new player handler
func NewPlayerHandler(
	sessionFactory session.SessionFactory,
	joinGameAction *action.JoinGameAction,
	getPlayerAction *query.GetPlayerAction,
) *PlayerHandler {
	return &PlayerHandler{
		BaseHandler:     NewBaseHandler(),
		sessionFactory:  sessionFactory,
		joinGameAction:  joinGameAction,
		getPlayerAction: getPlayerAction,
	}
}

// JoinGame adds a player to a game
func (h *PlayerHandler) JoinGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["gameId"]

	if gameID == "" {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Game ID is required")
		return
	}

	var req dto.JoinGameRequest
	if err := h.ParseJSONRequest(r, &req); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Use JoinGameAction (no playerID parameter for HTTP - it generates one)
	result, err := h.joinGameAction.Execute(r.Context(), gameID, req.PlayerName)
	if err != nil {
		h.logger.Error("Failed to join game", zap.Error(err),
			zap.String("game_id", gameID),
			zap.String("player_name", req.PlayerName))
		h.WriteErrorResponse(w, http.StatusBadRequest, "Failed to join game")
		return
	}

	// Convert to DTO and respond
	response := dto.JoinGameResponse{
		Game:     result.GameDto,
		PlayerID: result.PlayerID,
	}

	h.WriteJSONResponse(w, http.StatusOK, response)
}

// GetPlayer retrieves a player by ID
func (h *PlayerHandler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["gameId"]
	playerID := vars["playerId"]

	if gameID == "" {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Game ID is required")
		return
	}

	if playerID == "" {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Player ID is required")
		return
	}

	// Get session
	sess := h.sessionFactory.Get(gameID)
	if sess == nil {
		h.logger.Error("Game session not found", zap.String("game_id", gameID))
		h.WriteErrorResponse(w, http.StatusNotFound, "Game session not found")
		return
	}

	// Use GetPlayerAction
	player, err := h.getPlayerAction.Execute(r.Context(), sess, playerID)
	if err != nil {
		h.logger.Error("Failed to get player", zap.Error(err),
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
		h.WriteErrorResponse(w, http.StatusNotFound, "Player not found")
		return
	}

	// Convert to DTO and respond (dereference pointer to value)
	// Phase state now managed by Game, pass it to mapper (dereference Game pointer)
	playerDto := dto.ToPlayerDto(*sess.Game(), *player, nil)
	response := dto.GetPlayerResponse{
		Player: playerDto,
	}
	h.WriteJSONResponse(w, http.StatusOK, response)
}
