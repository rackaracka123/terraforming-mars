package http

import (
	"net/http"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GameHandler handles HTTP requests for games
type GameHandler struct {
	gameService *service.GameService
}

// NewGameHandler creates a new game handler
func NewGameHandler(gameService *service.GameService) *GameHandler {
	return &GameHandler{
		gameService: gameService,
	}
}

// Use DTOs from the dto package
type CreateGameRequest = dto.CreateGameRequest
type JoinGameRequest = dto.JoinGameRequest

// CreateGame handles POST /games
func (h *GameHandler) CreateGame(c *gin.Context) {
	log := logger.Get()
	
	var req CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid create game request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Convert request to domain settings
	settings := domain.GameSettings{
		MaxPlayers: req.MaxPlayers,
	}

	log.Info("Creating new game", zap.Int("max_players", req.MaxPlayers))

	// Create game
	game, err := h.gameService.CreateGame(settings)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	log.Info("Game created successfully", zap.String("game_id", game.ID))

	response := dto.CreateGameResponse{
		Game: dto.ToGameDto(game),
	}
	c.JSON(http.StatusCreated, response)
}

// GetGame handles GET /games/:id
func (h *GameHandler) GetGame(c *gin.Context) {
	gameID := c.Param("id")
	log := logger.WithGameContext(gameID, "")
	
	if gameID == "" {
		log.Warn("Get game request missing game ID")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "game ID is required",
		})
		return
	}

	log.Debug("Getting game")

	game, err := h.gameService.GetGame(gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	log.Debug("Game retrieved successfully")

	response := dto.GetGameResponse{
		Game: dto.ToGameDto(game),
	}
	c.JSON(http.StatusOK, response)
}

// JoinGame handles POST /games/:id/join
func (h *GameHandler) JoinGame(c *gin.Context) {
	gameID := c.Param("id")
	log := logger.WithGameContext(gameID, "")
	
	if gameID == "" {
		log.Warn("Join game request missing game ID")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "game ID is required",
		})
		return
	}

	var req JoinGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid join game request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	log.Info("Player joining game", zap.String("player_name", req.PlayerName))

	game, err := h.gameService.JoinGame(gameID, req.PlayerName)
	if err != nil {
		log.Error("Failed to join game", 
			zap.Error(err),
			zap.String("player_name", req.PlayerName),
		)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Find the player that was just added
	var playerID string
	for _, player := range game.Players {
		if player.Name == req.PlayerName {
			playerID = player.ID
			break
		}
	}

	log.Info("Player joined game successfully", 
		zap.String("player_name", req.PlayerName),
		zap.String("player_id", playerID),
	)

	response := dto.JoinGameResponse{
		Game:     dto.ToGameDto(game),
		PlayerID: playerID,
	}
	c.JSON(http.StatusOK, response)
}

// ListGames handles GET /games
func (h *GameHandler) ListGames(c *gin.Context) {
	status := c.Query("status")
	log := logger.Get()

	log.Debug("Listing games", zap.String("status_filter", status))

	games, err := h.gameService.ListGames(status)
	if err != nil {
		log.Error("Failed to list games", zap.Error(err), zap.String("status_filter", status))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	log.Debug("Games listed successfully", zap.Int("count", len(games)))

	response := dto.ListGamesResponse{
		Games: dto.ToGameDtoSlice(games),
	}
	c.JSON(http.StatusOK, response)
}
