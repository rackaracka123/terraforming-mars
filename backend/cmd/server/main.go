package main

import (
	"log"
	"net/http"
	"os"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	"terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize repositories
	helloRepo := repository.NewHelloRepository()

	// Initialize services
	helloService := service.NewHelloService(helloRepo)

	// Initialize handlers
	helloHandler := httpHandler.NewHelloHandler(helloService)

	// Initialize WebSocket hub
	hub := websocket.NewHub(helloService)
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
	r.GET("/health", helloHandler.HealthCheck)

	// Hello endpoint
	r.GET("/hello", helloHandler.GetHello)

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		hub.ServeWS(c.Writer, c.Request)
	})

	// Get port from environment or default to 3001
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("Hello World backend server starting on port %s", port)
	log.Printf("Health check available at: http://localhost:%s/health", port)
	log.Printf("Hello endpoint available at: http://localhost:%s/hello", port)
	log.Printf("WebSocket endpoint available at: ws://localhost:%s/ws", port)

	// Start server
	if err := r.Run(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
