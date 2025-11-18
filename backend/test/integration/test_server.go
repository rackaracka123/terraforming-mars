package integration

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"terraforming-mars-backend/internal/cards"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	wsHandler "terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/internal/session"
	sessionGame "terraforming-mars-backend/internal/session/game"
	sessionCard "terraforming-mars-backend/internal/session/game/card"
	sessionDeck "terraforming-mars-backend/internal/session/game/deck"
	sessionPlayer "terraforming-mars-backend/internal/session/game/player"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// TestServer represents a test server instance
type TestServer struct {
	server    *http.Server
	wsService *wsHandler.WebSocketService
	hubCtx    context.Context
	cancel    context.CancelFunc
	port      int
	logger    *zap.Logger
	started   bool
	mu        sync.Mutex
	// Repositories and event bus for state management
	gameRepo            *repository.GameRepositoryImpl
	playerRepo          *repository.PlayerRepositoryImpl
	cardDeckRepo        *repository.CardDeckRepositoryImpl
	eventBus            *events.EventBusImpl
	forcedActionManager cards.ForcedActionManager
}

// NewTestServer creates a new test server on the specified port
func NewTestServer(port int) (*TestServer, error) {
	logger := zap.NewNop() // Use no-op logger for tests to reduce noise

	// Initialize event bus
	eventBus := events.NewEventBus()

	// Initialize OLD repositories (still used by some services during migration)
	playerRepo := repository.NewPlayerRepository(eventBus).(*repository.PlayerRepositoryImpl)
	gameRepo := repository.NewGameRepository(eventBus).(*repository.GameRepositoryImpl)
	cardRepo := repository.NewCardRepository()

	// Load card data for integration testing
	if err := cardRepo.LoadCards(context.Background()); err != nil {
		logger.Warn("Failed to load card data in test server, using fallback", zap.Error(err))
	}

	cardDeckRepo := repository.NewCardDeckRepository().(*repository.CardDeckRepositoryImpl)

	// Initialize NEW session repositories
	newDeckRepo, err := sessionDeck.NewRepository(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize deck repository: %w", err)
	}
	newGameRepo := sessionGame.NewRepository(eventBus)
	newPlayerRepo := sessionPlayer.NewRepository(eventBus)
	newCardRepo := sessionCard.NewRepository(newDeckRepo, cardRepo)

	// Create Hub first
	hub := core.NewHub()

	// Create SessionManager with NEW repositories
	sessionManager := session.NewSessionManager(newGameRepo, newPlayerRepo, newCardRepo, hub, eventBus)

	// Create services with proper SessionManager dependency
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo, cardRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	forcedActionManager.SubscribeToPhaseChanges()
	playerService := service.NewPlayerService(gameRepo, playerRepo, sessionManager, boardService, tileService, forcedActionManager, eventBus)
	cardService := service.NewCardService(newGameRepo, newPlayerRepo, newCardRepo, cardDeckRepo, sessionManager, tileService, effectSubscriber, forcedActionManager)
	gameService := service.NewGameService(gameRepo, playerRepo, cardRepo, cardService, cardDeckRepo, boardService, sessionManager)
	standardProjectService := service.NewStandardProjectService(gameRepo, playerRepo, sessionManager, tileService)
	resourceConversionService := service.NewResourceConversionService(gameRepo, playerRepo, boardService, sessionManager, eventBus)
	adminService := service.NewAdminService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, effectSubscriber, forcedActionManager)

	// Register card-specific listeners (removed since we're using mock cards)
	// if err := initialization.RegisterCardListeners(eventBus); err != nil {
	// 	return nil, fmt.Errorf("failed to register card listeners: %w", err)
	// }

	// Initialize WebSocket service with Hub (use nil for actions in tests for now)
	wsService := wsHandler.NewWebSocketService(gameService, playerService, standardProjectService, cardService, adminService, resourceConversionService, gameRepo, playerRepo, cardRepo, hub, nil, nil, nil, nil, nil)

	// Setup router
	mainRouter := mux.NewRouter()
	apiRouter := httpHandler.SetupRouter(gameService, playerService, cardService, playerRepo, cardRepo, nil)
	mainRouter.PathPrefix("/api/v1").Handler(apiRouter)
	mainRouter.HandleFunc("/ws", wsService.ServeWS)

	// Add health check endpoint
	mainRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mainRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &TestServer{
		server:              server,
		wsService:           wsService,
		port:                port,
		logger:              logger,
		gameRepo:            gameRepo,
		playerRepo:          playerRepo,
		cardDeckRepo:        cardDeckRepo,
		eventBus:            eventBus,
		forcedActionManager: forcedActionManager,
	}, nil
}

// Start starts the test server
func (ts *TestServer) Start() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.started {
		return nil
	}

	// Start WebSocket service
	ts.hubCtx, ts.cancel = context.WithCancel(context.Background())
	go ts.wsService.Run(ts.hubCtx)

	// Start HTTP server in background
	go func() {
		if err := ts.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ts.logger.Error("Test server failed", zap.Error(err))
		}
	}()

	// Wait for server to be ready with polling mechanism
	if err := ts.waitForServerReady(); err != nil {
		return fmt.Errorf("server failed to start: %w", err)
	}

	ts.started = true
	return nil
}

// waitForServerReady polls the health endpoint until server is ready
func (ts *TestServer) waitForServerReady() error {
	healthURL := fmt.Sprintf("http://localhost:%d/health", ts.port)

	// Try for up to 5 seconds with exponential backoff
	maxAttempts := 15
	baseDelay := 50 * time.Millisecond

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Use a fresh client for each attempt with shorter timeout
		client := &http.Client{Timeout: 200 * time.Millisecond}
		resp, err := client.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		// Exponential backoff with cap
		delay := time.Duration(1<<uint(attempt)) * baseDelay
		if delay > 500*time.Millisecond {
			delay = 500 * time.Millisecond
		}
		time.Sleep(delay)
	}

	return fmt.Errorf("server did not become ready within timeout on port %d", ts.port)
}

// Stop stops the test server with proper cleanup
func (ts *TestServer) Stop() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.started {
		return nil
	}

	// Stop WebSocket hub first to prevent new connections
	if ts.cancel != nil {
		ts.cancel()
	}

	// Stop HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ts.started = false

	// Shutdown server
	if err := ts.server.Shutdown(ctx); err != nil {
		// Force close if graceful shutdown fails
		ts.server.Close()
		return err
	}

	return nil
}

// GetBaseURL returns the base URL for HTTP requests
func (ts *TestServer) GetBaseURL() string {
	return fmt.Sprintf("http://localhost:%d", ts.port)
}

// GetWebSocketURL returns the WebSocket URL
func (ts *TestServer) GetWebSocketURL() string {
	return fmt.Sprintf("ws://localhost:%d/ws", ts.port)
}

// ClearState clears all repository and event bus state for test isolation
func (ts *TestServer) ClearState() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Clear all repositories
	if ts.gameRepo != nil {
		ts.gameRepo.Clear()
	}
	if ts.playerRepo != nil {
		ts.playerRepo.Clear()
	}
	if ts.cardDeckRepo != nil {
		ts.cardDeckRepo.Clear()
	}

	// Clear event bus subscriptions and re-subscribe system handlers
	if ts.eventBus != nil {
		ts.eventBus.Clear()
		// Re-subscribe ForcedActionManager to phase changes
		if ts.forcedActionManager != nil {
			ts.forcedActionManager.SubscribeToPhaseChanges()
		}
	}

	// Clear WebSocket connections to prevent old connections from interfering
	if ts.wsService != nil {
		hub := ts.wsService.GetHub()
		if hub != nil {
			hub.ClearConnections()
		}
	}
}
