package http

import (
	"net/http"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
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

// Use DTOs from the dto package
type CreateGameRequest = dto.CreateGameRequest
type JoinGameRequest = dto.JoinGameRequest

// CreateGame handles POST /games
func (h *GameHandler) CreateGame(c *gin.Context) {
	var req CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Convert request to domain settings
	settings := domain.GameSettings{
		MaxPlayers: req.MaxPlayers,
	}

	// Create game
	game, err := h.gameService.CreateGame(settings)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	response := dto.CreateGameResponse{
		Game: dto.ToGameDto(game),
	}
	c.JSON(http.StatusCreated, response)
}

// GetGame handles GET /games/:id
func (h *GameHandler) GetGame(c *gin.Context) {
	gameID := c.Param("id")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "game ID is required",
		})
		return
	}

	game, err := h.gameService.GetGame(gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	response := dto.GetGameResponse{
		Game: dto.ToGameDto(game),
	}
	c.JSON(http.StatusOK, response)
}

// JoinGame handles POST /games/:id/join
func (h *GameHandler) JoinGame(c *gin.Context) {
	gameID := c.Param("id")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "game ID is required",
		})
		return
	}

	var req JoinGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	game, err := h.gameService.JoinGame(gameID, req.PlayerName)
	if err != nil {
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

	response := dto.JoinGameResponse{
		Game:     dto.ToGameDto(game),
		PlayerID: playerID,
	}
	c.JSON(http.StatusOK, response)
}

// ListGames handles GET /games
func (h *GameHandler) ListGames(c *gin.Context) {
	status := c.Query("status")

	games, err := h.gameService.ListGames(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	response := dto.ListGamesResponse{
		Games: dto.ToGameDtoSlice(games),
	}
	c.JSON(http.StatusOK, response)
}
