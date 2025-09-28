package integration

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	wsHandler "terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
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
}

// NewTestServer creates a new test server on the specified port
func NewTestServer(port int) (*TestServer, error) {
	logger := zap.NewNop() // Use no-op logger for tests to reduce noise

	// Initialize repositories
	playerRepo := repository.NewPlayerRepository()
	gameRepo := repository.NewGameRepository()

	// Initialize services with proper event bus wiring
	cardRepo := repository.NewCardRepository()

	// Load card data for integration testing
	if err := cardRepo.LoadCards(context.Background()); err != nil {
		logger.Warn("Failed to load card data in test server, using fallback", zap.Error(err))
	}

	cardDeckRepo := repository.NewCardDeckRepository()

	// Create Hub first
	hub := core.NewHub()

	// Create SessionManager with Hub
	sessionManager := session.NewSessionManager(gameRepo, playerRepo, cardRepo, hub)

	// Create services with proper SessionManager dependency
	playerService := service.NewPlayerService(gameRepo, playerRepo, sessionManager)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager)
	boardService := service.NewBoardService()
	gameService := service.NewGameService(gameRepo, playerRepo, cardService.(*service.CardServiceImpl), boardService, sessionManager)
	standardProjectService := service.NewStandardProjectService(gameRepo, playerRepo, sessionManager)

	// Register card-specific listeners (removed since we're using mock cards)
	// if err := initialization.RegisterCardListeners(eventBus); err != nil {
	// 	return nil, fmt.Errorf("failed to register card listeners: %w", err)
	// }

	// Initialize WebSocket service with Hub
	wsService := wsHandler.NewWebSocketService(gameService, playerService, standardProjectService, cardService, gameRepo, playerRepo, cardRepo, hub)

	// Setup router
	mainRouter := mux.NewRouter()
	apiRouter := httpHandler.SetupRouter(gameService, playerService, cardService)
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
		server:    server,
		wsService: wsService,
		port:      port,
		logger:    logger,
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
