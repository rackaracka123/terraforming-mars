package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"
)

// ConnectPlayerAction handles player connection and reconnection logic
// This action encapsulates all business logic for connecting a player to a game
type ConnectPlayerAction struct {
	lobbyService   lobby.Service
	playerRepo     player.Repository
	sessionManager session.SessionManager
}

// ConnectResult contains the result of a connection action
type ConnectResult struct {
	PlayerID string
	GameID   string
	IsNew    bool
}

// NewConnectPlayerAction creates a new connect player action
func NewConnectPlayerAction(
	lobbyService lobby.Service,
	playerRepo player.Repository,
	sessionManager session.SessionManager,
) *ConnectPlayerAction {
	return &ConnectPlayerAction{
		lobbyService:   lobbyService,
		playerRepo:     playerRepo,
		sessionManager: sessionManager,
	}
}

// Execute handles both new connections and reconnections
// Returns ConnectResult containing the player ID, game ID, and whether this is a new connection
func (a *ConnectPlayerAction) Execute(ctx context.Context, gameID string, playerName string, requestedPlayerID string) (*ConnectResult, error) {
	log := logger.WithGameContext(gameID, requestedPlayerID)
	log.Debug("🔗 Executing connect player action",
		zap.String("player_name", playerName),
		zap.String("requested_player_id", requestedPlayerID))

	// Validate game exists
	_, err := a.lobbyService.GetGame(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for connection", zap.Error(err))
		return nil, fmt.Errorf("game not found: %w", err)
	}

	// Check if player already exists - prioritize ID for reconnection
	var playerID string
	if requestedPlayerID != "" {
		// Try to find by ID first (for reconnection)
		playerID = a.findExistingPlayerByID(ctx, gameID, requestedPlayerID)
		if playerID != "" {
			log.Debug("🔄 Found player by ID - treating as reconnection",
				zap.String("player_id", playerID),
				zap.String("player_name", playerName))
		}
	}

	// Only check by name if no playerID was provided or found
	// This prevents creating duplicate players when reconnecting
	if playerID == "" && requestedPlayerID == "" {
		// Fall back to finding by name only for new connections
		playerID = a.findExistingPlayerByName(ctx, gameID, playerName)
	}

	// Determine if this is a new connection or reconnection
	isNew := playerID == ""

	if isNew {
		return a.handleNewConnection(ctx, gameID, playerName)
	}

	return a.handleReconnection(ctx, gameID, playerID, playerName)
}

// handleNewConnection creates a new player and joins them to the game
func (a *ConnectPlayerAction) handleNewConnection(ctx context.Context, gameID string, playerName string) (*ConnectResult, error) {
	log := logger.WithGameContext(gameID, "")
	log.Debug("✨ Handling new player connection", zap.String("player_name", playerName))

	// Generate a new player ID
	playerID := uuid.New().String()

	log.Debug("🆕 Generated new player ID",
		zap.String("player_id", playerID))

	// Join game using the pre-generated player ID
	_, err := a.lobbyService.JoinGameWithPlayerID(ctx, gameID, playerName, playerID)
	if err != nil {
		log.Error("Failed to join game", zap.Error(err))
		return nil, fmt.Errorf("failed to join game: %w", err)
	}

	log.Info("✅ Player connected and joined game",
		zap.String("player_id", playerID),
		zap.String("player_name", playerName))

	return &ConnectResult{
		PlayerID: playerID,
		GameID:   gameID,
		IsNew:    true,
	}, nil
}

// handleReconnection reconnects an existing player to the game
func (a *ConnectPlayerAction) handleReconnection(ctx context.Context, gameID string, playerID string, playerName string) (*ConnectResult, error) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🔄 Handling existing player reconnection",
		zap.String("player_id", playerID),
		zap.String("player_name", playerName))

	// Update player connection status
	err := a.playerRepo.UpdateConnectionStatus(ctx, gameID, playerID, true)
	if err != nil {
		log.Error("Failed to update player connection status", zap.Error(err))
		// Continue with reconnection even if status update fails
	}

	// Broadcast complete game state to all players
	err = a.sessionManager.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast game state for reconnection", zap.Error(err))
		return nil, fmt.Errorf("failed to broadcast game state: %w", err)
	}

	log.Info("✅ Player reconnection completed successfully",
		zap.String("player_id", playerID),
		zap.String("player_name", playerName))

	return &ConnectResult{
		PlayerID: playerID,
		GameID:   gameID,
		IsNew:    false,
	}, nil
}

// findExistingPlayerByName checks if a player with the given name exists in the game
func (a *ConnectPlayerAction) findExistingPlayerByName(ctx context.Context, gameID, playerName string) string {
	// Get all players in the game
	players, err := a.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return ""
	}

	// Search for player by name
	for _, p := range players {
		if p.Name == playerName {
			return p.ID
		}
	}

	return ""
}

// findExistingPlayerByID checks if a player with the given ID exists in the game
func (a *ConnectPlayerAction) findExistingPlayerByID(ctx context.Context, gameID, playerID string) string {
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return ""
	}
	return player.ID
}

// SendConfirmation sends a connection confirmation message to the player
// This should be called AFTER connection.SetPlayer() has been called
func (a *ConnectPlayerAction) SendConfirmation(ctx context.Context, result *ConnectResult) error {
	log := logger.WithGameContext(result.GameID, result.PlayerID)
	log.Debug("📤 Sending connection confirmation to player")

	// Send personalized game state to the connected player
	// This serves as the connection confirmation
	err := a.sessionManager.Send(result.GameID, result.PlayerID)
	if err != nil {
		log.Error("Failed to send connection confirmation", zap.Error(err))
		return fmt.Errorf("failed to send confirmation: %w", err)
	}

	log.Info("✅ Connection confirmation sent successfully",
		zap.Bool("is_new", result.IsNew))
	return nil
}
