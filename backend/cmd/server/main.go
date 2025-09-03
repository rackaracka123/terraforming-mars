package main

import (
	"log"
	"net/http"
	"os"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize repositories
	gameRepo := repository.NewGameRepository()

	// Initialize services
	gameService := service.NewGameService(gameRepo)

	// Initialize handlers
	gameHandler := httpHandler.NewGameHandler(gameService)

	// Initialize Gin router
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Game endpoints
	r.POST("/games", gameHandler.CreateGame)
	r.GET("/games", gameHandler.ListGames)
	r.GET("/games/available", gameHandler.GetAvailableGames)
	r.GET("/games/:id", gameHandler.GetGame)
	r.POST("/games/:id/join", gameHandler.JoinGame)

	// Get port from environment or default to 3001
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("Terraforming Mars backend server starting on port %s", port)
	log.Printf("Health check available at: http://localhost:%s/health", port)
	log.Printf("Game endpoints available at: http://localhost:%s/games", port)

	// Start server
	if err := r.Run(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
