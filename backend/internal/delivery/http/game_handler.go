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

// GameHandler handles HTTP requests related to game operations
type GameHandler struct {
	*BaseHandler
	appStore *store.Store
}

// NewGameHandler creates a new game handler
func NewGameHandler(appStore *store.Store) *GameHandler {
	return &GameHandler{
		BaseHandler: NewBaseHandler(),
		appStore:    appStore,
	}
}

// CreateGame creates a new game
func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateGameRequest
	if err := h.ParseJSONRequest(r, &req); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create action to dispatch to store
	gameSettings := model.GameSettings{
		MaxPlayers: req.MaxPlayers,
	}

	gameID := generateGameID() // Generate a new game ID
	action := store.CreateGameAction(gameID, gameSettings, "http")

	// Dispatch to store
	if err := h.appStore.Dispatch(r.Context(), action); err != nil {
		logger.Error("Failed to create game", zap.Error(err))
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create game")
		return
	}

	// Get the created game from store
	game, exists := h.appStore.GetGame(gameID)
	if !exists {
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create game")
		return
	}

	// Convert to DTO and respond
	gameDto := dto.ToGameDtoBasic(game.Game())
	response := dto.CreateGameResponse{
		Game: gameDto,
	}
	h.WriteJSONResponse(w, http.StatusCreated, response)
}

// generateGameID generates a unique game ID
func generateGameID() string {
	return uuid.New().String()
}

// GetGame retrieves a game by ID
func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["gameId"]

	if gameID == "" {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Game ID is required")
		return
	}

	// Get game from store
	state := h.appStore.GetState()
	gameState, exists := state.GetGame(gameID)
	if !exists {
		logger.Error("Failed to get game", zap.String("game_id", gameID), zap.String("error", "game not found"))
		h.WriteErrorResponse(w, http.StatusNotFound, "Game not found")
		return
	}

	// Convert to DTO and respond
	gameDto := dto.ToGameDtoBasic(gameState.Game())
	response := dto.GetGameResponse{
		Game: gameDto,
	}
	h.WriteJSONResponse(w, http.StatusOK, response)
}

// ListGames retrieves all games
func (h *GameHandler) ListGames(w http.ResponseWriter, r *http.Request) {
	// Get all games from store
	state := h.appStore.GetState()

	// Convert gameStates to model.Game slice
	stateGames := state.Games()
	games := make([]model.Game, 0, len(stateGames))
	for _, gameState := range stateGames {
		games = append(games, gameState.Game())
	}

	// Convert to DTOs and respond
	gameDtos := dto.ToGameDtoSlice(games)
	response := dto.ListGamesResponse{
		Games: gameDtos,
	}
	h.WriteJSONResponse(w, http.StatusOK, response)
}
