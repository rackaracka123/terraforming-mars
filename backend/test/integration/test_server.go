package integration

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	wsHandler "terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/initialization"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// TestServer represents a test server instance
type TestServer struct {
	server  *http.Server
	hub     *wsHandler.Hub
	hubCtx  context.Context
	cancel  context.CancelFunc
	port    int
	logger  *zap.Logger
	started bool
	mu      sync.Mutex
}

// NewTestServer creates a new test server on the specified port
func NewTestServer(port int) (*TestServer, error) {
	logger := zap.NewNop() // Use no-op logger for tests to reduce noise

	// Initialize event bus
	eventBus := events.NewInMemoryEventBus()

	// Initialize repositories
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	parametersRepo := repository.NewGlobalParametersRepository(eventBus)

	// Initialize services with proper event bus wiring
	cardService := service.NewCardService(gameRepo, playerRepo)
	gameService := service.NewGameService(gameRepo, playerRepo, parametersRepo, cardService.(*service.CardServiceImpl), eventBus)
	playerService := service.NewPlayerService(gameRepo, playerRepo)
	globalParametersService := service.NewGlobalParametersService(gameRepo, parametersRepo)
	standardProjectService := service.NewStandardProjectService(gameRepo, playerRepo, parametersRepo, globalParametersService)

	// Register card-specific listeners
	if err := initialization.RegisterCardListeners(eventBus); err != nil {
		return nil, fmt.Errorf("failed to register card listeners: %w", err)
	}

	// Initialize WebSocket hub with proper event bus
	hub := wsHandler.NewHub(gameService, playerService, globalParametersService, standardProjectService, cardService, eventBus)

	wsHandlerInstance := wsHandler.NewHandler(hub)

	// Setup router
	mainRouter := mux.NewRouter()
	apiRouter := httpHandler.SetupRouter(gameService, playerService)
	mainRouter.PathPrefix("/api/v1").Handler(apiRouter)
	mainRouter.HandleFunc("/ws", wsHandlerInstance.ServeWS)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mainRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &TestServer{
		server: server,
		hub:    hub,
		port:   port,
		logger: logger,
	}, nil
}

// Start starts the test server
func (ts *TestServer) Start() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.started {
		return nil
	}

	// Start WebSocket hub
	ts.hubCtx, ts.cancel = context.WithCancel(context.Background())
	go ts.hub.Run(ts.hubCtx)

	// Start HTTP server in background
	go func() {
		if err := ts.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ts.logger.Error("Test server failed", zap.Error(err))
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)
	ts.started = true

	return nil
}

// Stop stops the test server
func (ts *TestServer) Stop() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.started {
		return nil
	}

	// Stop WebSocket hub
	if ts.cancel != nil {
		ts.cancel()
	}

	// Stop HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ts.started = false
	return ts.server.Shutdown(ctx)
}

// GetBaseURL returns the base URL for HTTP requests
func (ts *TestServer) GetBaseURL() string {
	return fmt.Sprintf("http://localhost:%d", ts.port)
}

// GetWebSocketURL returns the WebSocket URL
func (ts *TestServer) GetWebSocketURL() string {
	return fmt.Sprintf("ws://localhost:%d/ws", ts.port)
}