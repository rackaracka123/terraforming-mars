package http

import (
	"net/http"
	"terraforming-mars-backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

// GameHandler handles HTTP requests for game operations
type GameHandler struct {
	gameUC *usecase.GameUseCase
}

// NewGameHandler creates a new game handler
func NewGameHandler(gameUC *usecase.GameUseCase) *GameHandler {
	return &GameHandler{
		gameUC: gameUC,
	}
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the server is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *GameHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": c.GetTime("timestamp"),
	})
}

// GetGame godoc
// @Summary Get game by ID
// @Description Retrieve a game state by its ID
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "Game ID"
// @Success 200 {object} domain.GameState
// @Failure 404 {object} map[string]string
// @Router /games/{id} [get]
func (h *GameHandler) GetGame(c *gin.Context) {
	gameID := c.Param("id")
	
	game, err := h.gameUC.GetGame(gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}
	
	c.JSON(http.StatusOK, game)
}

// GetAvailableCorporations godoc
// @Summary Get available corporations
// @Description Get the list of corporations available for selection
// @Tags corporations
// @Accept json
// @Produce json
// @Success 200 {array} domain.Corporation
// @Router /corporations [get]
func (h *GameHandler) GetAvailableCorporations(c *gin.Context) {
	corporations := h.gameUC.GetAvailableCorporations()
	c.JSON(http.StatusOK, corporations)
}