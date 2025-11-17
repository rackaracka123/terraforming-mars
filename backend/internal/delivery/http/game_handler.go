package http

import (
	"net/http"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/internal/session/game"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// GameHandler handles HTTP requests related to game operations
type GameHandler struct {
	*BaseHandler
	gameService      service.GameService
	playerRepo       repository.PlayerRepository
	cardRepo         repository.CardRepository
	createGameAction *action.CreateGameAction
}

// NewGameHandler creates a new game handler
func NewGameHandler(gameService service.GameService, playerRepo repository.PlayerRepository, cardRepo repository.CardRepository, createGameAction *action.CreateGameAction) *GameHandler {
	return &GameHandler{
		BaseHandler:      NewBaseHandler(),
		gameService:      gameService,
		playerRepo:       playerRepo,
		cardRepo:         cardRepo,
		createGameAction: createGameAction,
	}
}

// CreateGame creates a new game
func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateGameRequest
	if err := h.ParseJSONRequest(r, &req); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert to subdomain settings
	gameSettings := game.GameSettings{
		MaxPlayers:      req.MaxPlayers,
		DevelopmentMode: req.DevelopmentMode,
	}

	// Use new action pattern
	createdGame, err := h.createGameAction.Execute(r.Context(), gameSettings)
	if err != nil {
		h.logger.Error("Failed to create game", zap.Error(err))
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create game")
		return
	}

	// Convert to model.Game for DTO
	modelGame := convertToModelGame(createdGame)

	// Convert to DTO and respond
	gameDto := dto.ToGameDtoBasic(modelGame, dto.GetPaymentConstants())
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
	games, err := h.gameService.ListGames(r.Context(), "")
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

// convertToModelGame converts a game.Game to model.Game for DTO compatibility
func convertToModelGame(g *game.Game) model.Game {
	return model.Game{
		ID:        g.ID,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
		Status:    model.GameStatus(g.Status),
		Settings: model.GameSettings{
			MaxPlayers:      g.Settings.MaxPlayers,
			Temperature:     g.Settings.Temperature,
			Oxygen:          g.Settings.Oxygen,
			Oceans:          g.Settings.Oceans,
			DevelopmentMode: g.Settings.DevelopmentMode,
			CardPacks:       g.Settings.CardPacks,
		},
		PlayerIDs:        g.PlayerIDs,
		HostPlayerID:     g.HostPlayerID,
		CurrentPhase:     model.GamePhase(g.CurrentPhase),
		GlobalParameters: g.GlobalParameters,
		ViewingPlayerID:  g.ViewingPlayerID,
		CurrentTurn:      g.CurrentTurn,
		Generation:       g.Generation,
		Board:            g.Board,
	}
}
