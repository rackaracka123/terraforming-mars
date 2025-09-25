package session

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/delivery/dto"
	apperrors "terraforming-mars-backend/internal/errors"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// SessionManager manages WebSocket sessions and provides broadcasting capabilities
// This service is used by services to broadcast complete game state to players
type SessionManager interface {
	// Session management - one session (connection) per player per game
	RegisterSession(playerID, gameID string, sendMessage func(dto.WebSocketMessage))
	UnregisterSession(playerID, gameID string)

	// Core broadcasting operations - both send complete game state with all data
	Broadcast(gameID string) error             // Send complete game state to all players in game
	Send(gameID string, playerID string) error // Send complete game state to specific player
}

// SessionManagerImpl implements the SessionManager interface
type SessionManagerImpl struct {
	// Session storage: game -> player -> connection
	sessions map[string]map[string]func(dto.WebSocketMessage)

	// Dependencies for game state broadcasting
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
	cardRepo   repository.CardRepository

	// Synchronization
	mu     sync.RWMutex
	logger *zap.Logger
}

// NewSessionManager creates a new session manager
func NewSessionManager(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	cardRepo repository.CardRepository,
) SessionManager {
	return &SessionManagerImpl{
		sessions:   make(map[string]map[string]func(dto.WebSocketMessage)),
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
		cardRepo:   cardRepo,
		logger:     logger.Get(),
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
		var notFoundErr *apperrors.NotFoundError
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

	// Debug: Check if players have starting card data
	for _, player := range players {
		log.Debug("🔍 Player data from repository",
			zap.String("player_id", player.ID),
			zap.String("player_name", player.Name),
			zap.Int("starting_selection_count", len(player.StartingSelection)),
			zap.Strings("starting_card_ids", player.StartingSelection))
	}

	// Fetch cards for players once (shared data)
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
		personalizedGameDTO := dto.ToGameDto(game, players, player.ID, playerCards, startingCards)

		// Send game state via direct session call
		err = sm.sendToPlayerDirect(player.ID, gameID, dto.MessageTypeGameUpdated, dto.GameUpdatedPayload{
			Game: personalizedGameDTO,
		})
		if err != nil {
			// Handle missing sessions gracefully - this can happen during test cleanup
			var notFoundErr *apperrors.NotFoundError
			var sessionErr *apperrors.SessionNotFoundError
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
			card, err := sm.cardRepo.GetCardByID(ctx, cardID)
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
			card, err := sm.cardRepo.GetCardByID(ctx, cardID)
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

// sendToPlayerDirect sends a message directly to a specific player's session
func (sm *SessionManagerImpl) sendToPlayerDirect(playerID, gameID string, messageType dto.MessageType, payload any) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	gameMap, exists := sm.sessions[gameID]
	if !exists {
		return &apperrors.SessionNotFoundError{Resource: "game", ID: gameID}
	}

	sendFunc, exists := gameMap[playerID]
	if !exists {
		return &apperrors.SessionNotFoundError{Resource: "player", ID: playerID}
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
