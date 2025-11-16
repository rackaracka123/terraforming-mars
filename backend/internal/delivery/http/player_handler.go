package http

import (
	"net/http"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/player"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// PlayerHandler handles HTTP requests related to player operations
type PlayerHandler struct {
	*BaseHandler
	playerRepo   player.Repository
	lobbyService lobby.Service
}

// NewPlayerHandler creates a new player handler
func NewPlayerHandler(playerRepo player.Repository, lobbyService lobby.Service) *PlayerHandler {
	return &PlayerHandler{
		BaseHandler:  NewBaseHandler(),
		playerRepo:   playerRepo,
		lobbyService: lobbyService,
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

	// Delegate to service
	game, err := h.lobbyService.JoinGame(r.Context(), gameID, req.PlayerName)
	if err != nil {
		h.logger.Error("Failed to join game", zap.Error(err),
			zap.String("game_id", gameID),
			zap.String("player_name", req.PlayerName))
		h.WriteErrorResponse(w, http.StatusBadRequest, "Failed to join game")
		return
	}

	// Find the player ID of the newly joined player - it's the last one added
	var playerID string
	if len(game.PlayerIDs) > 0 {
		// The newly joined player is the last one in the list
		playerID = game.PlayerIDs[len(game.PlayerIDs)-1]
	}

	// Convert to DTO and respond
	response := dto.JoinGameResponse{
		Game:     dto.ToGameDtoLobbyOnly(game, dto.GetPaymentConstants()),
		PlayerID: playerID,
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

	// Get player from repository
	player, err := h.playerRepo.GetByID(r.Context(), gameID, playerID)
	if err != nil {
		h.logger.Error("Failed to get player", zap.Error(err),
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
		h.WriteErrorResponse(w, http.StatusNotFound, "Player not found")
		return
	}

	// Convert to DTO and respond
	playerDto, err := dto.ToPlayerDto(player, nil)
	if err != nil {
		h.logger.Error("Failed to convert player to DTO", zap.Error(err),
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to convert player")
		return
	}

	response := dto.GetPlayerResponse{
		Player: playerDto,
	}
	h.WriteJSONResponse(w, http.StatusOK, response)
}
