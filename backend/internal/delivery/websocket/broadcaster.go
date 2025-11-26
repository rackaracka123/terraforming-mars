package websocket

import (
	"context"
	"errors"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game/board"
	"terraforming-mars-backend/internal/session/game/card"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// SessionBroadcaster handles broadcasting game state updates via WebSocket for a SPECIFIC game
// This component is bound to a single gameID and can never broadcast to a different game
// This component lives in the delivery layer and knows about DTOs and WebSocket Hub
type SessionBroadcaster struct {
	gameID         string // IMMUTABLE - bound at creation, cannot change
	gameRepo       game.Repository
	sessionFactory session.SessionFactory
	cardRepo       card.Repository
	boardRepo      board.Repository
	hub            *core.Hub
	logger         *zap.Logger
}

// NewSessionBroadcaster creates a new session-aware broadcaster bound to a specific game
func NewSessionBroadcaster(
	gameID string,
	gameRepo game.Repository,
	sessionFactory session.SessionFactory,
	cardRepo card.Repository,
	boardRepo board.Repository,
	hub *core.Hub,
) session.SessionManager {
	return &SessionBroadcaster{
		gameID:         gameID, // Permanently bound to this game
		gameRepo:       gameRepo,
		sessionFactory: sessionFactory,
		cardRepo:       cardRepo,
		boardRepo:      boardRepo,
		hub:            hub,
		logger:         logger.Get(),
	}
}

// GetGameID returns the game ID this broadcaster is bound to
func (b *SessionBroadcaster) GetGameID() string {
	return b.gameID
}

// Broadcast sends complete game state to all players in THIS game
func (b *SessionBroadcaster) Broadcast() error {
	ctx := context.Background()
	return b.broadcastGameStateInternal(ctx, "")
}

// Send sends complete game state to a specific player in THIS game
func (b *SessionBroadcaster) Send(playerID string) error {
	ctx := context.Background()
	return b.broadcastGameStateInternal(ctx, playerID)
}

// broadcastGameStateInternal gathers all game data, creates personalized DTOs, and broadcasts to all players
// If playerID is empty, broadcasts to all players; if specified, sends only to that player
// Uses b.gameID which is immutably bound at creation time
func (b *SessionBroadcaster) broadcastGameStateInternal(ctx context.Context, targetPlayerID string) error {
	gameID := b.gameID // Use the immutable game ID this broadcaster is bound to
	log := b.logger.With(zap.String("game_id", gameID))
	if targetPlayerID == "" {
		log.Info("🚀 Broadcasting game state to all players")
	} else {
		log.Info("🚀 Sending game state to specific player", zap.String("target_player_id", targetPlayerID))
	}

	// Get updated game state from repository
	newGame, err := b.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		// Handle missing game gracefully - this can happen during test cleanup or if game was deleted
		var notFoundErr *types.NotFoundError
		if errors.As(err, &notFoundErr) {
			log.Debug("Game no longer exists, skipping broadcast", zap.Error(err))
			return nil // No error, just skip the broadcast
		}
		log.Error("Failed to get game for broadcast", zap.Error(err))
		return err
	}

	// Get board from board repository and attach to game
	sessionBoard, err := b.boardRepo.Get(ctx)
	if err != nil {
		// Board may not exist for games created before board migration
		// Log warning but continue with empty board
		log.Warn("⚠️  Failed to get board for game, using empty board", zap.Error(err))
	} else {
		// Attach board to game
		newGame.Board = *sessionBoard
		log.Debug("🗺️  Fetched board for game", zap.Int("tile_count", len(sessionBoard.Tiles)))
	}

	// Get all players for personalized game states from session
	sess := b.sessionFactory.Get(gameID)
	if sess == nil {
		log.Error("Game session not found for broadcast")
		return fmt.Errorf("game session not found: %s", gameID)
	}

	sessionPlayers := sess.GetAllPlayers()

	// Convert player types to model types for DTO compatibility
	players := playersToModel(sessionPlayers)

	// If no players exist, there's nothing to broadcast
	if len(players) == 0 {
		log.Debug("No players found for game, skipping broadcast")
		return nil
	}

	log.Debug("📢 Broadcasting game state",
		zap.String("game_id", gameID),
		zap.Int("player_count", len(players)))

	for _, player := range players {
		log.Debug("🔍 Player state in broadcast",
			zap.String("player_id", player.ID),
			zap.String("player_name", player.Name),
			zap.Strings("cards_in_hand", player.Cards),
			zap.Int("hand_size", len(player.Cards)),
			zap.Strings("played_cards", player.PlayedCards),
			zap.Int("played_count", len(player.PlayedCards)))
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
			// Collect available starting cards
			for _, cardID := range player.SelectStartingCardsPhase.AvailableCards {
				allCardIds[cardID] = struct{}{}
			}
			log.Debug("Added starting cards to resolution",
				zap.Int("card_count", len(player.SelectStartingCardsPhase.AvailableCards)))

			// CRITICAL: Add corporation cards to resolved cards for frontend display
			for _, corpID := range player.SelectStartingCardsPhase.AvailableCorporations {
				allCardIds[corpID] = struct{}{}
			}
			log.Debug("Added corporations to resolution",
				zap.Int("corporation_count", len(player.SelectStartingCardsPhase.AvailableCorporations)),
				zap.Strings("corporation_ids", player.SelectStartingCardsPhase.AvailableCorporations))
		}
		// Add cards from PendingCardSelection (card selection effects)
		if player.PendingCardSelection != nil {
			for _, cardID := range player.PendingCardSelection.AvailableCards {
				allCardIds[cardID] = struct{}{}
			}
			log.Debug("Added pending card selection cards to resolution",
				zap.Int("card_count", len(player.PendingCardSelection.AvailableCards)))
		}
		// Add cards from PendingCardDrawSelection (card draw/peek/take/buy effects)
		if player.PendingCardDrawSelection != nil {
			for _, cardID := range player.PendingCardDrawSelection.AvailableCards {
				allCardIds[cardID] = struct{}{}
			}
			log.Debug("Added pending card draw selection cards to resolution",
				zap.Int("card_count", len(player.PendingCardDrawSelection.AvailableCards)))
		}
	}

	newResolvedCards, err := b.cardRepo.ListCardsByIdMap(ctx, allCardIds)
	if err != nil {
		log.Error("Failed to resolve card data for broadcast", zap.Error(err))
		return err
	}

	// Convert card types to model types for DTO compatibility
	resolvedCards := cardsToModel(newResolvedCards)

	log.Debug("Resolved cards for broadcast",
		zap.Int("total_card_ids", len(allCardIds)),
		zap.Int("resolved_cards", len(resolvedCards)))

	// Filter players based on target
	playersToSend := players
	if targetPlayerID != "" {
		// Find specific player
		playersToSend = []types.Player{}
		for _, player := range players {
			if player.ID == targetPlayerID {
				playersToSend = []types.Player{player}
				break
			}
		}
		if len(playersToSend) == 0 {
			return fmt.Errorf("target player %s not found in game %s", targetPlayerID, gameID)
		}
	}

	// Get payment constants once for all players
	paymentConstants := dto.GetPaymentConstants()

	// Send personalized game state to target player(s)
	for _, player := range playersToSend {
		log.Debug("🎯 Creating personalized DTO",
			zap.String("viewing_player_id", player.ID),
			zap.String("viewing_player_name", player.Name),
			zap.Int("total_players", len(players)))

		for i, p := range players {
			log.Debug("🔍 Player in array",
				zap.Int("index", i),
				zap.String("player_id", p.ID),
				zap.String("player_name", p.Name))
		}

		personalizedGameDTO := dto.ToGameDto(*newGame, players, player.ID, resolvedCards, paymentConstants)

		// Send game state via Hub
		err = b.sendToPlayerDirect(player.ID, gameID, dto.MessageTypeGameUpdated, dto.GameUpdatedPayload{
			Game: personalizedGameDTO,
		})
		if err != nil {
			// Handle missing sessions gracefully - this can happen during test cleanup
			var notFoundErr *types.NotFoundError
			var sessionErr *types.SessionNotFoundError
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
func (b *SessionBroadcaster) sendToPlayerDirect(playerID, gameID string, messageType dto.MessageType, payload any) error {
	message := dto.WebSocketMessage{
		Type:    messageType,
		Payload: payload,
		GameID:  gameID,
	}

	// Use the Hub to send the message
	err := b.hub.SendToPlayer(gameID, playerID, message)
	if err != nil {
		return err
	}

	b.logger.Debug("💬 Message sent to player via Hub",
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.String("message_type", string(messageType)))

	return nil
}

// ========== Type Converters: Session types → Model types ==========

// gameToModel converts a game.Game pointer to a types.Game value
func gameToModel(g *game.Game) types.Game {
	return *g
}

// playersToModel converts a slice of player.Player pointers to a slice of types.Player values
func playersToModel(players []*player.Player) []types.Player {
	result := make([]types.Player, len(players))
	for i, p := range players {
		// Extract the embedded types.Player from the wrapped player
		result[i] = *p.Player
	}
	return result
}

// cardsToModel converts a map of card.Card values to a map of types.Card values
func cardsToModel(cards map[string]card.Card) map[string]types.Card {
	return cards
}

// boardToModel converts a board.Board pointer to a types.Board value
func boardToModel(b *board.Board) types.Board {
	if b == nil {
		return types.Board{Tiles: []types.Tile{}}
	}

	tiles := make([]types.Tile, len(b.Tiles))
	for i, tile := range b.Tiles {
		// Convert bonuses
		bonuses := make([]types.TileBonus, len(tile.Bonuses))
		for j, bonus := range tile.Bonuses {
			bonuses[j] = types.TileBonus{
				Type:   types.ResourceType(bonus.Type),
				Amount: bonus.Amount,
			}
		}

		// Convert occupant if exists
		var occupant *types.TileOccupant
		if tile.OccupiedBy != nil {
			tags := make([]string, len(tile.OccupiedBy.Tags))
			copy(tags, tile.OccupiedBy.Tags)
			occupant = &types.TileOccupant{
				Type: types.ResourceType(tile.OccupiedBy.Type),
				Tags: tags,
			}
		}

		// Convert tile
		tiles[i] = types.Tile{
			Coordinates: types.HexPosition{
				Q: tile.Coordinates.Q,
				R: tile.Coordinates.R,
				S: tile.Coordinates.S,
			},
			Tags:        tile.Tags,
			Type:        types.ResourceType(tile.Type),
			Location:    types.TileLocation(tile.Location),
			DisplayName: tile.DisplayName,
			Bonuses:     bonuses,
			OccupiedBy:  occupant,
			OwnerID:     tile.OwnerID,
		}
	}

	return types.Board{Tiles: tiles}
}
