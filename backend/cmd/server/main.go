package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/initialization"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	wsHandler "terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/listeners"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/middleware"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	if err := logger.Init(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Shutdown()

	log := logger.Get()
	log.Info("Starting Terraforming Mars backend server")

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Initialize repositories
	gameRepo := repository.NewGameRepository()
	log.Info("Game repository initialized")

	// Initialize event system
	eventBus := events.NewInMemoryEventBus()
	log.Info("Event bus initialized")

	// Initialize card registry
	cardRegistry := cards.NewCardHandlerRegistry()
	if err := initialization.RegisterCardsWithRegistry(cardRegistry); err != nil {
		log.Fatal("Failed to register card handlers", zap.Error(err))
	}
	log.Info("Card registry initialized with handlers", zap.Int("handlers", len(cardRegistry.GetAllRegisteredCards())))

	// Initialize services
	gameService := service.NewGameService(gameRepo, eventBus, cardRegistry)
	log.Info("Game service initialized")

	// Register event listeners
	listenerRegistry := listeners.NewRegistry(eventBus, gameRepo, cardRegistry)
	listenerRegistry.RegisterAllListeners()
	log.Info("Event listeners registered")
	
	// Register card-specific listeners
	if err := initialization.RegisterCardListeners(eventBus); err != nil {
		log.Fatal("Failed to register card listeners", zap.Error(err))
	}
	log.Info("Card listeners registered")

	// Initialize handlers
	gameHandler := httpHandler.NewGameHandler(gameService)
	log.Info("HTTP handlers initialized")

	// Initialize WebSocket hub
	wsHub := wsHandler.NewHub(gameService)
	go wsHub.Run() // Start hub in a goroutine
	log.Info("WebSocket hub started")

	// Initialize Gin router
	r := gin.New() // Use gin.New() instead of gin.Default() to have full control

	// Add middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.ZapLogger())
	r.Use(middleware.ZapRecovery())

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
	r.GET("/games/:id", gameHandler.GetGame)
	r.POST("/games/:id/join", gameHandler.JoinGame)

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		wsHandler.ServeWS(wsHub, c.Writer, c.Request)
	})

	// Get port from environment or default to 3001
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Info("Server configuration",
		zap.String("port", port),
		zap.String("health_endpoint", "http://localhost:"+port+"/health"),
		zap.String("games_endpoint", "http://localhost:"+port+"/games"),
		zap.String("websocket_endpoint", "ws://localhost:"+port+"/ws"),
	)

	// Start server in a goroutine
	go func() {
		log.Info("Starting HTTP server", zap.String("port", port))
		if err := r.Run(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-quit
	log.Info("Shutting down server gracefully...")
	log.Info("Server stopped")
}
