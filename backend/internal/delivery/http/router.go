package http

import (
	"net/http"
	httpmiddleware "terraforming-mars-backend/internal/middleware/http"
	"terraforming-mars-backend/internal/store"

	"github.com/gorilla/mux"
)

// SetupRouter creates and configures the HTTP router
func SetupRouter(appStore *store.Store) *mux.Router {
	// Create handlers
	gameHandler := NewGameHandler(appStore)
	playerHandler := NewPlayerHandler(appStore)
	healthHandler := NewHealthHandler()

	// Create router
	router := mux.NewRouter()

	// Apply middleware
	router.Use(httpmiddleware.Recovery)
	router.Use(httpmiddleware.CORS)
	router.Use(httpmiddleware.LoggingMiddleware)
	router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", healthHandler.HealthCheck).Methods(http.MethodGet)

	// Game routes
	gameRoutes := api.PathPrefix("/games").Subrouter()
	gameRoutes.HandleFunc("", gameHandler.CreateGame).Methods(http.MethodPost)
	gameRoutes.HandleFunc("", gameHandler.ListGames).Methods(http.MethodGet)
	gameRoutes.HandleFunc("/{gameId}", gameHandler.GetGame).Methods(http.MethodGet)

	// Player routes
	playerRoutes := api.PathPrefix("/games/{gameId}/players").Subrouter()
	playerRoutes.HandleFunc("", playerHandler.JoinGame).Methods(http.MethodPost)
	playerRoutes.HandleFunc("/{playerId}", playerHandler.GetPlayer).Methods(http.MethodGet)
	playerRoutes.HandleFunc("/{playerId}/resources", playerHandler.UpdatePlayerResources).Methods(http.MethodPut)


	return router
}
