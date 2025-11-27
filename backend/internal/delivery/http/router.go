package http

import (
	"net/http"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/action/query"
	httpmiddleware "terraforming-mars-backend/internal/middleware/http"

	"github.com/gorilla/mux"
)

// SetupRouter creates HTTP router
// Includes both query (GET) and mutation (POST) endpoints
func SetupRouter(
	createGameAction *action.CreateGameAction,
	getGameAction *query.GetGameAction,
	listGamesAction *query.ListGamesAction,
	getPlayerAction *query.GetPlayerAction,
) *mux.Router {
	// Create handlers
	gameHandler := NewGameHandler(createGameAction, getGameAction, listGamesAction)
	playerHandler := NewPlayerHandler(getPlayerAction)
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

	// Player routes (query only)
	playerRoutes := api.PathPrefix("/games/{gameId}/players").Subrouter()
	playerRoutes.HandleFunc("/{playerId}", playerHandler.GetPlayer).Methods(http.MethodGet)

	return router
}
