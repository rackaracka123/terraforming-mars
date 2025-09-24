package session

import (
	"context"
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// CardService interface for card retrieval (to avoid import cycle)
type CardService interface {
	GetCardByID(ctx context.Context, cardID string) (*model.Card, error)
}

// SessionManager manages WebSocket sessions and provides broadcasting capabilities
// This service is used by services to broadcast messages to players
type SessionManager interface {
	// Session management - one session (connection) per player per game
	RegisterSession(playerID, gameID string, sendMessage func(dto.WebSocketMessage))
	UnregisterSession(playerID, gameID string)

	// Core broadcasting operations
	BroadcastToGame(gameID string, messageType dto.MessageType, payload any) error
	BroadcastToGameExcept(gameID string, messageType dto.MessageType, payload any, excludePlayerID string) error
	SendToPlayer(playerID, gameID string, messageType dto.MessageType, payload any) error

	// High-level game state broadcasting - gathers data and sends personalized game states
	BroadcastGameState(ctx context.Context, gameID string) error
}

// SessionManagerImpl implements the SessionManager interface
type SessionManagerImpl struct {
	// Session storage: game -> player -> connection
	sessions map[string]map[string]func(dto.WebSocketMessage)

	// Dependencies for game state broadcasting
	gameRepo    repository.GameRepository
	playerRepo  repository.PlayerRepository
	cardService CardService

	// Synchronization
	mu     sync.RWMutex
	logger *zap.Logger
}

// NewSessionManager creates a new session manager
func NewSessionManager(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	cardService CardService,
) SessionManager {
	return &SessionManagerImpl{
		sessions:    make(map[string]map[string]func(dto.WebSocketMessage)),
		gameRepo:    gameRepo,
		playerRepo:  playerRepo,
		cardService: cardService,
		logger:      logger.Get(),
	}
}

// RegisterSession registers a new session
func (sm *SessionManagerImpl) RegisterSession(playerID, gameID string, sendMessage func(dto.WebSocketMessage)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Initialize game map if needed
	if sm.sessions[gameID] == nil {
		sm.sessions[gameID] = make(map[string]func(dto.WebSocketMessage))
	}

	// Register the session (replaces any existing session for this player)
	sm.sessions[gameID][playerID] = sendMessage

	sm.logger.Debug("✅ Session registered",
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}

// UnregisterSession removes a session
func (sm *SessionManagerImpl) UnregisterSession(playerID, gameID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if gameMap, exists := sm.sessions[gameID]; exists {
		delete(gameMap, playerID)
		// Clean up empty game map
		if len(gameMap) == 0 {
			delete(sm.sessions, gameID)
		}
	}

	sm.logger.Debug("❌ Session unregistered",
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}

// BroadcastToGame broadcasts a message to all players in a game
func (sm *SessionManagerImpl) BroadcastToGame(gameID string, messageType dto.MessageType, payload any) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	gameMap, exists := sm.sessions[gameID]
	if !exists || len(gameMap) == 0 {
		sm.logger.Debug("No sessions to broadcast to",
			zap.String("game_id", gameID),
			zap.String("message_type", string(messageType)))
		return nil
	}

	message := dto.WebSocketMessage{
		Type:    messageType,
		Payload: payload,
		GameID:  gameID,
	}

	sm.logger.Debug("📢 Broadcasting to game",
		zap.String("game_id", gameID),
		zap.String("message_type", string(messageType)),
		zap.Int("player_count", len(gameMap)))

	var lastError error
	successCount := 0

	for playerID, sendFunc := range gameMap {
		sendFunc(message)
		successCount++
		sm.logger.Debug("💬 Message sent to player",
			zap.String("player_id", playerID),
			zap.String("game_id", gameID),
			zap.String("message_type", string(messageType)))
	}

	sm.logger.Debug("📢 Broadcast completed",
		zap.String("game_id", gameID),
		zap.Int("successful_sends", successCount),
		zap.String("message_type", string(messageType)))

	return lastError
}

// BroadcastToGameExcept broadcasts a message to all players in a game except one
func (sm *SessionManagerImpl) BroadcastToGameExcept(gameID string, messageType dto.MessageType, payload any, excludePlayerID string) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	gameMap, exists := sm.sessions[gameID]
	if !exists || len(gameMap) == 0 {
		return nil
	}

	message := dto.WebSocketMessage{
		Type:    messageType,
		Payload: payload,
		GameID:  gameID,
	}

	sm.logger.Debug("📢 Broadcasting to game (except player)",
		zap.String("game_id", gameID),
		zap.String("exclude_player_id", excludePlayerID),
		zap.String("message_type", string(messageType)))

	var lastError error
	successCount := 0

	for playerID, sendFunc := range gameMap {
		if playerID != excludePlayerID {
			sendFunc(message)
			successCount++
		}
	}

	sm.logger.Debug("📢 Broadcast completed (except player)",
		zap.String("game_id", gameID),
		zap.Int("successful_sends", successCount),
		zap.String("message_type", string(messageType)))

	return lastError
}

// SendToPlayer sends a message to a specific player
func (sm *SessionManagerImpl) SendToPlayer(playerID, gameID string, messageType dto.MessageType, payload any) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	gameMap, exists := sm.sessions[gameID]
	if !exists {
		return fmt.Errorf("game %s not found", gameID)
	}

	sendFunc, exists := gameMap[playerID]
	if !exists {
		return fmt.Errorf("player %s not found in game %s", playerID, gameID)
	}

	message := dto.WebSocketMessage{
		Type:    messageType,
		Payload: payload,
		GameID:  gameID,
	}

	sendFunc(message)

	sm.logger.Debug("💬 Message sent to player",
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.String("message_type", string(messageType)))

	return nil
}

// BroadcastGameState gathers all game data, creates personalized DTOs, and broadcasts to all players
func (sm *SessionManagerImpl) BroadcastGameState(ctx context.Context, gameID string) error {
	log := sm.logger.With(zap.String("game_id", gameID))
	log.Info("🚀 Broadcasting game state to all players")

	// Get updated game state
	game, err := sm.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for broadcast", zap.Error(err))
		return err
	}

	// Get all players for personalized game states
	players, err := sm.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players for broadcast", zap.Error(err))
		return err
	}

	// Debug: Check if players have starting card data
	for _, player := range players {
		log.Debug("🔍 Player data from repository",
			zap.String("player_id", player.ID),
			zap.String("player_name", player.Name),
			zap.Int("starting_selection_count", len(player.StartingSelection)),
			zap.Strings("starting_card_ids", player.StartingSelection))
	}

	// Send personalized game state to each player
	for _, player := range players {
		// Fetch cards for players and create personalized game state for this player
		playerCards, err := sm.fetchPlayerCards(ctx, players)
		if err != nil {
			log.Warn("Failed to fetch player cards, using empty cards", zap.Error(err))
			playerCards = make(map[string][]model.Card)
		}
		startingCards, err := sm.fetchPlayerStartingCards(ctx, gameID, players)
		if err != nil {
			log.Warn("Failed to fetch player starting cards, using empty cards", zap.Error(err))
			startingCards = make(map[string][]model.Card)
		}
		personalizedGameDTO := dto.ToGameDto(game, players, player.ID, playerCards, startingCards)

		// Send updated game state with personalized view
		err = sm.SendToPlayer(player.ID, gameID, dto.MessageTypeGameUpdated, dto.GameUpdatedPayload{
			Game: personalizedGameDTO,
		})
		if err != nil {
			log.Error("Failed to send game state update to player",
				zap.Error(err),
				zap.String("player_id", player.ID))
			continue // Continue with other players
		}

		log.Debug("✅ Sent personalized game state to player",
			zap.String("player_id", player.ID))
	}

	log.Info("✅ Game state broadcast completed")
	return nil
}

// fetchPlayerCards retrieves card data for all players' hand cards
func (sm *SessionManagerImpl) fetchPlayerCards(ctx context.Context, players []model.Player) (map[string][]model.Card, error) {
	playerCards := make(map[string][]model.Card)

	for _, player := range players {
		if len(player.Cards) == 0 {
			playerCards[player.ID] = []model.Card{}
			continue
		}

		// Fetch cards for this player
		cards := make([]model.Card, 0, len(player.Cards))
		for _, cardID := range player.Cards {
			card, err := sm.cardService.GetCardByID(ctx, cardID)
			if err != nil {
				// Log warning but continue with other cards
				sm.logger.Warn("Failed to fetch card", zap.String("card_id", cardID), zap.Error(err))
				continue
			}
			if card != nil {
				cards = append(cards, *card)
			}
		}
		playerCards[player.ID] = cards
	}

	return playerCards, nil
}

// fetchPlayerStartingCards retrieves card data for all players' starting card selections
func (sm *SessionManagerImpl) fetchPlayerStartingCards(ctx context.Context, gameID string, players []model.Player) (map[string][]model.Card, error) {
	playerStartingCards := make(map[string][]model.Card)

	for _, player := range players {
		// Fetch fresh player data from repository to get latest starting card data
		freshPlayer, err := sm.playerRepo.GetByID(ctx, gameID, player.ID)
		if err != nil {
			sm.logger.Warn("Failed to fetch fresh player data, using cached data",
				zap.String("player_id", player.ID),
				zap.Error(err))
			freshPlayer = player // Fall back to cached data
		} else {
			sm.logger.Debug("🔄 Fetched fresh player data for starting cards",
				zap.String("player_id", player.ID),
				zap.Int("cached_starting_cards", len(player.StartingSelection)),
				zap.Int("fresh_starting_cards", len(freshPlayer.StartingSelection)))
		}

		sm.logger.Debug("🔍 Checking player starting cards",
			zap.String("player_id", freshPlayer.ID),
			zap.Int("starting_selection_count", len(freshPlayer.StartingSelection)),
			zap.Strings("starting_card_ids", freshPlayer.StartingSelection))

		if len(freshPlayer.StartingSelection) == 0 {
			sm.logger.Debug("⚠️ Player has no starting cards to fetch",
				zap.String("player_id", freshPlayer.ID))
			playerStartingCards[freshPlayer.ID] = []model.Card{}
			continue
		}

		// Fetch starting cards for this player
		cards := make([]model.Card, 0, len(freshPlayer.StartingSelection))
		for _, cardID := range freshPlayer.StartingSelection {
			card, err := sm.cardService.GetCardByID(ctx, cardID)
			if err != nil {
				// Log warning but continue with other cards
				sm.logger.Warn("Failed to fetch starting card", zap.String("card_id", cardID), zap.Error(err))
				continue
			}
			if card != nil {
				cards = append(cards, *card)
			}
		}
		playerStartingCards[freshPlayer.ID] = cards
	}

	return playerStartingCards, nil
}
