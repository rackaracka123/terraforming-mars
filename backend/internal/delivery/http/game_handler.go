package http

import (
	"net/http"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/service"

	"github.com/gin-gonic/gin"
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

// CreateGameRequest represents the request body for creating a game
type CreateGameRequest struct {
	MaxPlayers int `json:"maxPlayers" binding:"required,min=1,max=5"`
}

// JoinGameRequest represents the request body for joining a game
type JoinGameRequest struct {
	PlayerName string `json:"playerName" binding:"required,min=1,max=50"`
}

// CreateGame handles POST /games
func (h *GameHandler) CreateGame(c *gin.Context) {
	var req CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert request to domain settings
	settings := domain.GameSettings{
		MaxPlayers: req.MaxPlayers,
	}

	// Create game
	game, err := h.gameService.CreateGame(settings)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, game)
}

// GetGame handles GET /games/:id
func (h *GameHandler) GetGame(c *gin.Context) {
	gameID := c.Param("id")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game ID is required"})
		return
	}

	game, err := h.gameService.GetGame(gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, game)
}

// JoinGame handles POST /games/:id/join
func (h *GameHandler) JoinGame(c *gin.Context) {
	gameID := c.Param("id")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game ID is required"})
		return
	}

	var req JoinGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.gameService.JoinGame(gameID, req.PlayerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, game)
}

// ListGames handles GET /games
func (h *GameHandler) ListGames(c *gin.Context) {
	status := c.Query("status")

	games, err := h.gameService.ListGames(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"games": games})
}

