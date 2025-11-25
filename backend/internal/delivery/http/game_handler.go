package http

import (
	"net/http"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/action/query"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/types"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// GameHandler handles HTTP requests related to game operations
type GameHandler struct {
	*BaseHandler
	sessionFactory   session.SessionFactory
	createGameAction *action.CreateGameAction
	getGameAction    *query.GetGameAction
	listGamesAction  *query.ListGamesAction
}

// NewGameHandler creates a new game handler
func NewGameHandler(
	sessionFactory session.SessionFactory,
	createGameAction *action.CreateGameAction,
	getGameAction *query.GetGameAction,
	listGamesAction *query.ListGamesAction,
) *GameHandler {
	return &GameHandler{
		BaseHandler:      NewBaseHandler(),
		sessionFactory:   sessionFactory,
		createGameAction: createGameAction,
		getGameAction:    getGameAction,
		listGamesAction:  listGamesAction,
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

	// Convert to types.Game for DTO
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

	// Get session
	sess := h.sessionFactory.Get(gameID)
	if sess == nil {
		h.logger.Error("Game session not found", zap.String("game_id", gameID))
		h.WriteErrorResponse(w, http.StatusNotFound, "Game session not found")
		return
	}

	// Use GetGameAction
	result, err := h.getGameAction.Execute(r.Context(), sess, playerID)
	if err != nil {
		h.logger.Error("Failed to get game", zap.Error(err), zap.String("game_id", gameID))
		h.WriteErrorResponse(w, http.StatusNotFound, "Game not found")
		return
	}

	// Convert to types.Game for DTO compatibility
	game := convertToModelGame(result.Game)

	var gameDto dto.GameDto
	if playerID != "" && result.Players != nil {
		// Return personalized view with full player data
		gameDto = dto.ToGameDto(game, result.Players, playerID, result.ResolvedCards, dto.GetPaymentConstants())
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
	// Use ListGamesAction (list all games by passing empty status)
	games, err := h.listGamesAction.Execute(r.Context(), "")
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

// convertToModelGame converts a game.Game to types.Game for DTO compatibility
func convertToModelGame(g *game.Game) types.Game {
	return types.Game{
		ID:        g.ID,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
		Status:    types.GameStatus(g.Status),
		Settings: types.GameSettings{
			MaxPlayers:      g.Settings.MaxPlayers,
			Temperature:     g.Settings.Temperature,
			Oxygen:          g.Settings.Oxygen,
			Oceans:          g.Settings.Oceans,
			DevelopmentMode: g.Settings.DevelopmentMode,
			CardPacks:       g.Settings.CardPacks,
		},
		PlayerIDs:        g.PlayerIDs,
		HostPlayerID:     g.HostPlayerID,
		CurrentPhase:     types.GamePhase(g.CurrentPhase),
		GlobalParameters: g.GlobalParameters,
		ViewingPlayerID:  g.ViewingPlayerID,
		CurrentTurn:      g.CurrentTurn,
		Generation:       g.Generation,
		Board:            g.Board,
	}
}
