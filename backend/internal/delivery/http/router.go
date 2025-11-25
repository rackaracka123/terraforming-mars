package http

import (
	"net/http"
	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/action/query"
	httpmiddleware "terraforming-mars-backend/internal/middleware/http"
	"terraforming-mars-backend/internal/session"

	"github.com/gorilla/mux"
)

// SetupRouter creates and configures the HTTP router
func SetupRouter(
	sessionFactory session.SessionFactory,
	createGameAction *action.CreateGameAction,
	joinGameAction *action.JoinGameAction,
	getGameAction *query.GetGameAction,
	listGamesAction *query.ListGamesAction,
	getPlayerAction *query.GetPlayerAction,
	listCardsAction *query.ListCardsAction,
	getCorporationsAction *query.GetCorporationsAction,
) *mux.Router {
	// Create handlers
	gameHandler := NewGameHandler(sessionFactory, createGameAction, getGameAction, listGamesAction)
	playerHandler := NewPlayerHandler(sessionFactory, joinGameAction, getPlayerAction)
	cardHandler := NewCardHandler(listCardsAction, getCorporationsAction)
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

	// Card routes
	cardRoutes := api.PathPrefix("/cards").Subrouter()
	cardRoutes.HandleFunc("", cardHandler.ListCards).Methods(http.MethodGet)

	// Corporation routes
	api.HandleFunc("/corporations", cardHandler.GetCorporations).Methods(http.MethodGet)

	return router
}
