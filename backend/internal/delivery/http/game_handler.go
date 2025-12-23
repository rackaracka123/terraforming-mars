package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/action/query"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// GameHandler handles HTTP requests for games
type GameHandler struct {
	createGameAction      *gameaction.CreateGameAction
	createDemoLobbyAction *gameaction.CreateDemoLobbyAction
	getGameAction         *query.GetGameAction
	listGamesAction       *query.ListGamesAction
	listCardsAction       *query.ListCardsAction
	cardRegistry          cards.CardRegistry
}

// NewGameHandler creates a new game handler
func NewGameHandler(
	createGameAction *gameaction.CreateGameAction,
	createDemoLobbyAction *gameaction.CreateDemoLobbyAction,
	getGameAction *query.GetGameAction,
	listGamesAction *query.ListGamesAction,
	listCardsAction *query.ListCardsAction,
	cardRegistry cards.CardRegistry,
) *GameHandler {
	return &GameHandler{
		createGameAction:      createGameAction,
		createDemoLobbyAction: createDemoLobbyAction,
		getGameAction:         getGameAction,
		listGamesAction:       listGamesAction,
		listCardsAction:       listCardsAction,
		cardRegistry:          cardRegistry,
	}
}

// GetGame handles GET /api/v1/games/{gameId}
func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	log := logger.Get()
	ctx := r.Context()

	// Extract gameId from URL
	vars := mux.Vars(r)
	gameID := vars["gameId"]

	log.Info("📡 HTTP GET /api/v1/games/:gameId", zap.String("game_id", gameID))

	// Execute query action
	game, err := h.getGameAction.Execute(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	// Convert to DTO (HTTP GET has no authenticated player, use first player as fallback)
	gameDto := dto.ToGameDto(game, h.cardRegistry, "")

	// Wrap in response structure
	response := dto.GetGameResponse{
		Game: gameDto,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Info("✅ Game retrieved successfully", zap.String("game_id", gameID))
}

// ListGames handles GET /api/v1/games
func (h *GameHandler) ListGames(w http.ResponseWriter, r *http.Request) {
	log := logger.Get()
	ctx := r.Context()

	log.Info("📡 HTTP GET /api/v1/games")

	// Execute query action (no status filter for now)
	games, err := h.listGamesAction.Execute(ctx, nil)
	if err != nil {
		log.Error("Failed to list games", zap.Error(err))
		http.Error(w, "Failed to list games", http.StatusInternalServerError)
		return
	}

	// Convert to DTOs (HTTP GET has no authenticated player, use first player as fallback)
	gameDtos := make([]dto.GameDto, len(games))
	for i, game := range games {
		gameDtos[i] = dto.ToGameDto(game, h.cardRegistry, "")
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(gameDtos); err != nil {
		log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Info("✅ Games listed successfully", zap.Int("count", len(games)))
}

// CreateGame handles POST /api/v1/games
func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	log := logger.Get()
	ctx := r.Context()

	log.Info("📡 HTTP POST /api/v1/games")

	// Parse request body
	var req dto.CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert to GameSettings
	settings := game.GameSettings{
		MaxPlayers:      req.MaxPlayers,
		DevelopmentMode: req.DevelopmentMode,
		CardPacks:       req.CardPacks,
	}

	// Execute create game action
	game, err := h.createGameAction.Execute(ctx, settings)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		http.Error(w, "Failed to create game", http.StatusInternalServerError)
		return
	}

	// Convert to DTO (HTTP POST has no authenticated player yet, use first player as fallback)
	gameDto := dto.ToGameDto(game, h.cardRegistry, "")

	// Wrap in response structure
	response := dto.CreateGameResponse{
		Game: gameDto,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Info("✅ Game created successfully", zap.String("game_id", game.ID()))
}

// ListCards handles GET /api/v1/cards
func (h *GameHandler) ListCards(w http.ResponseWriter, r *http.Request) {
	log := logger.Get()
	ctx := r.Context()

	log.Info("📡 HTTP GET /api/v1/cards")

	// Parse query parameters
	queryParams := r.URL.Query()
	offset := 0
	limit := 100 // Default limit

	if offsetParam := queryParams.Get("offset"); offsetParam != "" {
		var parsedOffset int
		if _, err := fmt.Sscanf(offsetParam, "%d", &parsedOffset); err == nil {
			offset = parsedOffset
		}
	}

	if limitParam := queryParams.Get("limit"); limitParam != "" {
		var parsedLimit int
		if _, err := fmt.Sscanf(limitParam, "%d", &parsedLimit); err == nil {
			limit = parsedLimit
		}
	}

	// Execute query action
	result, err := h.listCardsAction.Execute(ctx, offset, limit)
	if err != nil {
		log.Error("Failed to list cards", zap.Error(err))
		http.Error(w, "Failed to list cards", http.StatusInternalServerError)
		return
	}

	// Convert to DTOs
	cardDtos := make([]dto.CardDto, len(result.Cards))
	for i, card := range result.Cards {
		cardDtos[i] = dto.ToCardDto(card)
	}

	// Wrap in response structure
	response := dto.ListCardsResponse{
		Cards:      cardDtos,
		TotalCount: result.TotalCount,
		Offset:     result.Offset,
		Limit:      result.Limit,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Info("✅ Cards listed successfully", zap.Int("count", len(cardDtos)))
}

// CreateDemoLobby handles POST /api/v1/games/demo/lobby
func (h *GameHandler) CreateDemoLobby(w http.ResponseWriter, r *http.Request) {
	log := logger.Get()
	ctx := r.Context()

	log.Info("POST /api/v1/games/demo/lobby")

	var req dto.CreateDemoLobbyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	settings := gameaction.DemoLobbySettings{
		PlayerCount: req.PlayerCount,
		CardPacks:   req.CardPacks,
		PlayerName:  req.PlayerName,
	}

	result, err := h.createDemoLobbyAction.Execute(ctx, settings)
	if err != nil {
		log.Error("Failed to create demo lobby", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.CreateDemoLobbyResponse{
		Game:     result.GameDto,
		PlayerID: result.PlayerID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Info("Demo lobby created successfully", zap.String("game_id", result.GameDto.ID))
}
