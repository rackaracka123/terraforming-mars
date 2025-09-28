package session

import (
	"context"
	"errors"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// SessionManager manages WebSocket sessions and provides broadcasting capabilities
// This service is used by services to broadcast complete game state to players
type SessionManager interface {
	// Core broadcasting operations - both send complete game state with all data
	Broadcast(gameID string) error             // Send complete game state to all players in game
	Send(gameID string, playerID string) error // Send complete game state to specific player
}

// SessionManagerImpl implements the SessionManager interface
type SessionManagerImpl struct {
	// Dependencies for game state broadcasting
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
	cardRepo   repository.CardRepository
	hub        *core.Hub

	// Synchronization
	logger *zap.Logger
}

// NewSessionManager creates a new session manager
func NewSessionManager(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	cardRepo repository.CardRepository,
	hub *core.Hub,
) SessionManager {
	return &SessionManagerImpl{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
		cardRepo:   cardRepo,
		hub:        hub,
		logger:     logger.Get(),
	}
}

// SessionManager no longer manages sessions directly - Hub handles all connections

// Broadcast sends complete game state to all players in a game
func (sm *SessionManagerImpl) Broadcast(gameID string) error {
	ctx := context.Background() // Using background context for broadcast operations
	return sm.broadcastGameStateInternal(ctx, gameID, "")
}

// Send sends complete game state to a specific player
func (sm *SessionManagerImpl) Send(gameID string, playerID string) error {
	ctx := context.Background() // Using background context for send operations
	return sm.broadcastGameStateInternal(ctx, gameID, playerID)
}

// broadcastGameStateInternal gathers all game data, creates personalized DTOs, and broadcasts to all players
// If playerID is empty, broadcasts to all players; if specified, sends only to that player
func (sm *SessionManagerImpl) broadcastGameStateInternal(ctx context.Context, gameID string, targetPlayerID string) error {
	log := sm.logger.With(zap.String("game_id", gameID))
	if targetPlayerID == "" {
		log.Info("🚀 Broadcasting game state to all players")
	} else {
		log.Info("🚀 Sending game state to specific player", zap.String("target_player_id", targetPlayerID))
	}

	// Get updated game state
	game, err := sm.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		// Handle missing game gracefully - this can happen during test cleanup or if game was deleted
		var notFoundErr *model.NotFoundError
		if errors.As(err, &notFoundErr) {
			log.Debug("Game no longer exists, skipping broadcast", zap.Error(err))
			return nil // No error, just skip the broadcast
		}
		log.Error("Failed to get game for broadcast", zap.Error(err))
		return err
	}

	// Get all players for personalized game states
	players, err := sm.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players for broadcast", zap.Error(err))
		return err
	}

	// If no players exist, there's nothing to broadcast
	if len(players) == 0 {
		log.Debug("No players found for game, skipping broadcast")
		return nil
	}

	for _, player := range players {
		log.Debug("🔍 Player data from repository",
			zap.String("player_id", player.ID),
			zap.String("player_name", player.Name))
	}

	allCardIds := make(map[string]struct{})
	for _, player := range players {
		for _, cardID := range player.Cards {
			allCardIds[cardID] = struct{}{}
		}
		for _, cardID := range player.PlayedCards {
			allCardIds[cardID] = struct{}{}
		}
		if player.ProductionPhase != nil {
			for _, cardID := range player.ProductionPhase.AvailableCards {
				allCardIds[cardID] = struct{}{}
			}
		}
		if player.SelectStartingCardsPhase != nil {
			for _, cardID := range player.SelectStartingCardsPhase.AvailableCards {
				allCardIds[cardID] = struct{}{}
			}
		}
	}

	resolvedCards, err := sm.cardRepo.ListCardsByIdMap(ctx, allCardIds)
	if err != nil {
		log.Error("Failed to resolve card data for broadcast", zap.Error(err))
		return err
	}

	// Filter players based on target
	playersToSend := players
	if targetPlayerID != "" {
		// Find specific player
		playersToSend = []model.Player{}
		for _, player := range players {
			if player.ID == targetPlayerID {
				playersToSend = []model.Player{player}
				break
			}
		}
		if len(playersToSend) == 0 {
			return fmt.Errorf("target player %s not found in game %s", targetPlayerID, gameID)
		}
	}

	// Send personalized game state to target player(s)
	for _, player := range playersToSend {
		personalizedGameDTO := dto.ToGameDto(game, players, player.ID, resolvedCards)

		// Send game state via direct session call
		err = sm.sendToPlayerDirect(player.ID, gameID, dto.MessageTypeGameUpdated, dto.GameUpdatedPayload{
			Game: personalizedGameDTO,
		})
		if err != nil {
			// Handle missing sessions gracefully - this can happen during test cleanup
			var notFoundErr *model.NotFoundError
			var sessionErr *model.SessionNotFoundError
			if errors.As(err, &notFoundErr) || errors.As(err, &sessionErr) {
				log.Debug("Player session no longer exists, skipping broadcast",
					zap.Error(err),
					zap.String("player_id", player.ID))
			} else {
				log.Error("Failed to send game state update to player",
					zap.Error(err),
					zap.String("player_id", player.ID))
			}
			continue // Continue with other players
		}

		log.Debug("✅ Sent personalized game state to player",
			zap.String("player_id", player.ID))
	}

	log.Info("✅ Game state broadcast completed")
	return nil
}

// sendToPlayerDirect sends a message directly to a specific player via the Hub
func (sm *SessionManagerImpl) sendToPlayerDirect(playerID, gameID string, messageType dto.MessageType, payload any) error {
	message := dto.WebSocketMessage{
		Type:    messageType,
		Payload: payload,
		GameID:  gameID,
	}

	// Use the Hub to send the message
	err := sm.hub.SendToPlayer(gameID, playerID, message)
	if err != nil {
		return err
	}

	sm.logger.Debug("💬 Message sent to player via Hub",
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.String("message_type", string(messageType)))

	return nil
}
