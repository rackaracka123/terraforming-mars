package http

import (
	"net/http"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/store"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// PlayerHandler handles HTTP requests related to player operations
type PlayerHandler struct {
	*BaseHandler
	appStore *store.Store
}

// NewPlayerHandler creates a new player handler
func NewPlayerHandler(appStore *store.Store) *PlayerHandler {
	return &PlayerHandler{
		BaseHandler: NewBaseHandler(),
		appStore:    appStore,
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

	// Generate player ID and create join game action
	playerID := generatePlayerID()
	action := store.JoinGameAction(gameID, playerID, req.PlayerName, "http")

	// Dispatch to store
	if err := h.appStore.Dispatch(r.Context(), action); err != nil {
		logger.Error("Failed to join game", zap.Error(err),
			zap.String("game_id", gameID),
			zap.String("player_name", req.PlayerName))
		h.WriteErrorResponse(w, http.StatusBadRequest, "Failed to join game")
		return
	}

	// Get updated game state from store
	gameState, exists := h.appStore.GetGame(gameID)
	if !exists {
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Game not found after join")
		return
	}

	// Convert to DTO and respond
	response := dto.JoinGameResponse{
		Game:     dto.ToGameDtoBasic(gameState.Game()),
		PlayerID: playerID,
	}

	h.WriteJSONResponse(w, http.StatusOK, response)
}

// generatePlayerID generates a unique player ID
func generatePlayerID() string {
	return uuid.New().String()
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

	// Get player from store
	playerState, exists := h.appStore.GetPlayer(playerID)
	if !exists || playerState.GameID() != gameID {
		logger.Error("Failed to get player", zap.String("game_id", gameID), zap.String("player_id", playerID))
		h.WriteErrorResponse(w, http.StatusNotFound, "Player not found")
		return
	}

	// Convert to DTO and respond
	cardRegistry := h.appStore.GetState().CardRegistry()
	playerDto := dto.ToPlayerDto(playerState.Player(), cardRegistry)
	response := dto.GetPlayerResponse{
		Player: playerDto,
	}
	h.WriteJSONResponse(w, http.StatusOK, response)
}

// UpdatePlayerResources updates player resources (for testing/admin purposes)
func (h *PlayerHandler) UpdatePlayerResources(w http.ResponseWriter, r *http.Request) {
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

	// Create update resources action
	action := store.UpdateResourcesAction(gameID, playerID, resources, "http")

	// Dispatch to store
	if err := h.appStore.Dispatch(r.Context(), action); err != nil {
		logger.Error("Failed to update player resources", zap.Error(err),
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
		h.WriteErrorResponse(w, http.StatusBadRequest, "Failed to update player resources")
		return
	}

	// Get updated player state from store
	playerState, exists := h.appStore.GetPlayer(playerID)
	if !exists {
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Player not found after update")
		return
	}

	// Convert to DTO and respond
	cardRegistry := h.appStore.GetState().CardRegistry()
	playerDto := dto.ToPlayerDto(playerState.Player(), cardRegistry)
	response := dto.UpdatePlayerResourcesResponse{
		Player: playerDto,
	}
	h.WriteJSONResponse(w, http.StatusOK, response)
}
