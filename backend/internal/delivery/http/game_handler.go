package http

import (
	"net/http"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// GameHandler handles HTTP requests related to game operations
type GameHandler struct {
	*BaseHandler
	gameService service.GameService
}

// NewGameHandler creates a new game handler
func NewGameHandler(gameService service.GameService) *GameHandler {
	return &GameHandler{
		BaseHandler: NewBaseHandler(),
		gameService: gameService,
	}
}

// CreateGame creates a new game
func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "CreateGame")
	
	var req dto.CreateGameRequest
	if err := h.ParseJSONRequest(r, &req); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Delegate to service
	gameSettings := model.GameSettings{
		MaxPlayers: req.MaxPlayers,
	}
	
	game, err := h.gameService.CreateGame(r.Context(), gameSettings)
	if err != nil {
		h.logger.Error("Failed to create game", zap.Error(err))
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create game")
		return
	}
	
	// Convert to DTO and respond
	gameDto := dto.ToGameDto(game)
	response := dto.CreateGameResponse{
		Game: gameDto,
	}
	h.WriteJSONResponse(w, http.StatusCreated, response)
}

// GetGame retrieves a game by ID
func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "GetGame")
	
	vars := mux.Vars(r)
	gameID := vars["gameId"]
	
	if gameID == "" {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Game ID is required")
		return
	}
	
	// Delegate to service
	game, err := h.gameService.GetGame(r.Context(), gameID)
	if err != nil {
		h.logger.Error("Failed to get game", zap.Error(err), zap.String("game_id", gameID))
		h.WriteErrorResponse(w, http.StatusNotFound, "Game not found")
		return
	}
	
	// Convert to DTO and respond
	gameDto := dto.ToGameDto(game)
	response := dto.GetGameResponse{
		Game: gameDto,
	}
	h.WriteJSONResponse(w, http.StatusOK, response)
}

// ListGames retrieves all games
func (h *GameHandler) ListGames(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "ListGames")
	
	// Delegate to service (list all games by passing empty status)
	games, err := h.gameService.ListGames(r.Context(), "")
	if err != nil {
		h.logger.Error("Failed to list games", zap.Error(err))
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list games")
		return
	}
	
	// Convert to DTOs and respond
	gameDtos := dto.ToGameDtoSlice(games)
	response := dto.ListGamesResponse{
		Games: gameDtos,
	}
	h.WriteJSONResponse(w, http.StatusOK, response)
}

