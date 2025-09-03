// @title Terraforming Mars API
// @version 1.0
// @description Digital implementation of Terraforming Mars board game backend API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3001
// @BasePath /api/v1

// @schemes http https
package main

import (
	"log"
	"net/http"
	"os"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	"terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Initialize repositories
	gameRepo := repository.NewGameRepository()

	// Initialize use cases
	gameUC := usecase.NewGameUseCase(gameRepo)

	// Initialize handlers
	gameHandler := httpHandler.NewGameHandler(gameUC)

	// Initialize WebSocket hub
	hub := websocket.NewHub(gameUC)
	go hub.Run()

	// Initialize Gin router
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// Health check endpoint
	r.GET("/health", gameHandler.HealthCheck)

	// API v1 routes
	api := r.Group("/api/v1")
	{
		// Game routes
		api.GET("/games/:id", gameHandler.GetGame)
		
		// Corporation routes
		api.GET("/corporations", gameHandler.GetAvailableCorporations)
	}

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		hub.ServeWS(c.Writer, c.Request)
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Get port from environment or default to 3001
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("Terraforming Mars backend server starting on port %s", port)
	log.Printf("Health check available at: http://localhost:%s/health", port)
	log.Printf("API documentation available at: http://localhost:%s/swagger/index.html", port)
	log.Printf("WebSocket endpoint available at: ws://localhost:%s/ws", port)

	// Start server
	if err := r.Run(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}