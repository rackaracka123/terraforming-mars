package http

import (
	"net/http"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/features/card"
	gameModel "terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/player"
	sessionpkg "terraforming-mars-backend/internal/session"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// GameHandler handles HTTP requests related to game operations
type GameHandler struct {
	*BaseHandler
	gameRepo     gameModel.Repository
	lobbyService lobby.Service
	playerRepo   player.Repository
	cardRepo     card.CardRepository
	sessionRepo  sessionpkg.Repository
}

// NewGameHandler creates a new game handler
func NewGameHandler(gameRepo gameModel.Repository, lobbyService lobby.Service, playerRepo player.Repository, cardRepo card.CardRepository, sessionRepo sessionpkg.Repository) *GameHandler {
	return &GameHandler{
		BaseHandler:  NewBaseHandler(),
		gameRepo:     gameRepo,
		lobbyService: lobbyService,
		playerRepo:   playerRepo,
		cardRepo:     cardRepo,
		sessionRepo:  sessionRepo,
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
	gameSettings := gameModel.GameSettings{
		MaxPlayers:      req.MaxPlayers,
		DevelopmentMode: req.DevelopmentMode,
	}

	game, err := h.lobbyService.CreateGame(r.Context(), gameSettings)
	if err != nil {
		h.logger.Error("Failed to create game", zap.Error(err))
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create game")
		return
	}

	// Convert to DTO and respond (lobby view - no full game state needed)
	gameDto := dto.ToGameDtoLobbyOnly(game, dto.GetPaymentConstants())
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

	// Get game from repository
	game, err := h.gameRepo.GetByID(r.Context(), gameID)
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
			// Cards are now Card instances (Living Card Instance Pattern)
			for _, card := range player.PlayedCards {
				allCardIds[card.ID] = struct{}{}
			}
			for _, card := range player.Cards {
				allCardIds[card.ID] = struct{}{}
			}
		}

		resolvedCards, err := h.cardRepo.ListCardsByIdMap(r.Context(), allCardIds)
		if err != nil {
			h.logger.Error("Failed to resolve cards", zap.Error(err), zap.String("game_id", gameID))
			h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to resolve cards")
			return
		}

		// Get session if game is active (has feature services)
		if game.Status == gameModel.GameStatusActive {
			gameSession, err := h.sessionRepo.GetByID(r.Context(), gameID)
			if err != nil {
				h.logger.Error("Failed to get session", zap.Error(err), zap.String("game_id", gameID))
				h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get game session")
				return
			}
			gameDto, err = dto.ToGameDto(
				game,
				players,
				playerID,
				resolvedCards,
				dto.GetPaymentConstants(),
				gameSession.ParametersService,
				gameSession.BoardService,
			)
			if err != nil {
				h.logger.Error("Failed to convert game to DTO", zap.Error(err), zap.String("game_id", gameID))
				h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to convert game")
				return
			}
		} else {
			// Game in lobby - no session/feature services yet
			// Return DTO with nil/default values for runtime state
			h.logger.Warn("Game is in lobby state, cannot provide full game data via HTTP", zap.String("game_id", gameID))
			h.WriteErrorResponse(w, http.StatusBadRequest, "Game is not active yet")
			return
		}
	} else {
		// Return basic non-personalized view
		// Game in lobby - no session/feature services yet
		h.logger.Warn("Cannot provide basic game view - game needs to be active")
		h.WriteErrorResponse(w, http.StatusBadRequest, "Game is not active yet")
		return
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
	gameDtos := dto.ToGameDtoSliceLobbyOnly(games, dto.GetPaymentConstants())
	response := dto.ListGamesResponse{
		Games: gameDtos,
	}
	h.WriteJSONResponse(w, http.StatusOK, response)
}
