package http

import (
	"net/http"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// PlayerHandler handles HTTP requests related to player operations
type PlayerHandler struct {
	*BaseHandler
	playerService service.PlayerService
	gameService   service.GameService
}

// NewPlayerHandler creates a new player handler
func NewPlayerHandler(playerService service.PlayerService, gameService service.GameService) *PlayerHandler {
	return &PlayerHandler{
		BaseHandler:   NewBaseHandler(),
		playerService: playerService,
		gameService:   gameService,
	}
}

// JoinGame adds a player to a game
func (h *PlayerHandler) JoinGame(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "JoinGame")
	
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
	game, err := h.gameService.JoinGame(r.Context(), gameID, req.PlayerName)
	if err != nil {
		h.logger.Error("Failed to join game", zap.Error(err), 
			zap.String("game_id", gameID),
			zap.String("player_name", req.PlayerName))
		h.WriteErrorResponse(w, http.StatusBadRequest, "Failed to join game")
		return
	}
	
	// Find the player ID of the newly joined player
	var playerID string
	for _, player := range game.Players {
		if player.Name == req.PlayerName {
			playerID = player.ID
			break
		}
	}
	
	// Convert to DTO and respond
	response := dto.JoinGameResponse{
		Game:     dto.ToGameDto(game),
		PlayerID: playerID,
	}
	
	h.WriteJSONResponse(w, http.StatusOK, response)
}

// GetPlayer retrieves a player by ID
func (h *PlayerHandler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "GetPlayer")
	
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
	
	// Delegate to service
	player, err := h.playerService.GetPlayer(r.Context(), gameID, playerID)
	if err != nil {
		h.logger.Error("Failed to get player", zap.Error(err), 
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
		h.WriteErrorResponse(w, http.StatusNotFound, "Player not found")
		return
	}
	
	// Convert to DTO and respond
	playerDto := dto.ToPlayerDto(*player)
	response := dto.GetPlayerResponse{
		Player: playerDto,
	}
	h.WriteJSONResponse(w, http.StatusOK, response)
}

// UpdatePlayerResources updates player resources (for testing/admin purposes)
func (h *PlayerHandler) UpdatePlayerResources(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "UpdatePlayerResources")
	
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
	
	var req dto.UpdateResourcesRequest
	if err := h.ParseJSONRequest(r, &req); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Convert DTO to model
	resources := model.Resources{
		Credits:  req.Resources.Credits,
		Steel:    req.Resources.Steel,
		Titanium: req.Resources.Titanium,
		Plants:   req.Resources.Plants,
		Energy:   req.Resources.Energy,
		Heat:     req.Resources.Heat,
	}
	
	// Delegate to service
	err := h.playerService.UpdatePlayerResources(r.Context(), gameID, playerID, resources)
	if err != nil {
		h.logger.Error("Failed to update player resources", zap.Error(err),
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
		h.WriteErrorResponse(w, http.StatusBadRequest, "Failed to update player resources")
		return
	}
	
	// Get updated player state
	player, err := h.playerService.GetPlayer(r.Context(), gameID, playerID)
	if err != nil {
		h.logger.Error("Failed to get player after update", zap.Error(err),
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get updated player state")
		return
	}
	
	// Convert to DTO and respond
	playerDto := dto.ToPlayerDto(*player)
	response := dto.UpdatePlayerResourcesResponse{
		Player: playerDto,
	}
	h.WriteJSONResponse(w, http.StatusOK, response)
}