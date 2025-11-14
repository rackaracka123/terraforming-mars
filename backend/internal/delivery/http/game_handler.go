package http

import (
	"net/http"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"
	"terraforming-mars-backend/internal/service"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// GameHandler handles HTTP requests related to game operations
type GameHandler struct {
	*BaseHandler
	gameService  service.GameService
	lobbyService lobby.Service
	playerRepo   player.Repository
	cardRepo     game.CardRepository
}

// NewGameHandler creates a new game handler
func NewGameHandler(gameService service.GameService, lobbyService lobby.Service, playerRepo player.Repository, cardRepo game.CardRepository) *GameHandler {
	return &GameHandler{
		BaseHandler:  NewBaseHandler(),
		gameService:  gameService,
		lobbyService: lobbyService,
		playerRepo:   playerRepo,
		cardRepo:     cardRepo,
	}
}

// CreateGame creates a new game
func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateGameRequest
	if err := h.ParseJSONRequest(r, &req); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Delegate to service
	gameSettings := model.GameSettings{
		MaxPlayers:      req.MaxPlayers,
		DevelopmentMode: req.DevelopmentMode,
	}

	game, err := h.lobbyService.CreateGame(r.Context(), gameSettings)
	if err != nil {
		h.logger.Error("Failed to create game", zap.Error(err))
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create game")
		return
	}

	// Convert to DTO and respond
	gameDto := dto.ToGameDtoBasic(game, dto.GetPaymentConstants())
	response := dto.CreateGameResponse{
		Game: gameDto,
	}
	h.WriteJSONResponse(w, http.StatusCreated, response)
}

// GetGame retrieves a game by ID
func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["gameId"]

	if gameID == "" {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Game ID is required")
		return
	}

	// Check for optional playerId query parameter for personalized view
	playerID := r.URL.Query().Get("playerId")

	// Delegate to service
	game, err := h.gameService.GetGame(r.Context(), gameID)
	if err != nil {
		h.logger.Error("Failed to get game", zap.Error(err), zap.String("game_id", gameID))
		h.WriteErrorResponse(w, http.StatusNotFound, "Game not found")
		return
	}

	var gameDto dto.GameDto
	if playerID != "" {
		// Return personalized view with full player data
		players, err := h.playerRepo.ListByGameID(r.Context(), gameID)
		if err != nil {
			h.logger.Error("Failed to get players", zap.Error(err), zap.String("game_id", gameID))
			h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get players")
			return
		}

		// Collect all card IDs that need resolution
		allCardIds := make(map[string]struct{})
		for _, player := range players {
			if player.Corporation != nil {
				allCardIds[player.Corporation.ID] = struct{}{}
			}
			for _, cardID := range player.PlayedCards {
				allCardIds[cardID] = struct{}{}
			}
			for _, cardID := range player.Cards {
				allCardIds[cardID] = struct{}{}
			}
		}

		resolvedCards, err := h.cardRepo.ListCardsByIdMap(r.Context(), allCardIds)
		if err != nil {
			h.logger.Error("Failed to resolve cards", zap.Error(err), zap.String("game_id", gameID))
			h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to resolve cards")
			return
		}

		gameDto = dto.ToGameDto(game, players, playerID, resolvedCards, dto.GetPaymentConstants())
	} else {
		// Return basic non-personalized view
		gameDto = dto.ToGameDtoBasic(game, dto.GetPaymentConstants())
	}

	response := dto.GetGameResponse{
		Game: gameDto,
	}
	h.WriteJSONResponse(w, http.StatusOK, response)
}

// ListGames retrieves all games
func (h *GameHandler) ListGames(w http.ResponseWriter, r *http.Request) {
	// Delegate to service (list all games by passing empty status)
	games, err := h.lobbyService.ListGames(r.Context(), "")
	if err != nil {
		h.logger.Error("Failed to list games", zap.Error(err))
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list games")
		return
	}

	// Convert to DTOs and respond
	gameDtos := dto.ToGameDtoSlice(games, dto.GetPaymentConstants())
	response := dto.ListGamesResponse{
		Games: gameDtos,
	}
	h.WriteJSONResponse(w, http.StatusOK, response)
}
