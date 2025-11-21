package session

import (
	"context"
	"errors"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/session/board"
	"terraforming-mars-backend/internal/session/card"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

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
	// Dependencies for game state broadcasting (NEW repository types)
	gameRepo   game.Repository
	playerRepo player.Repository
	cardRepo   card.Repository
	boardRepo  board.Repository
	hub        *core.Hub

	// Synchronization
	logger *zap.Logger
}

// NewSessionManager creates a new session manager and subscribes to domain events
func NewSessionManager(
	gameRepo game.Repository,
	playerRepo player.Repository,
	cardRepo card.Repository,
	boardRepo board.Repository,
	hub *core.Hub,
	eventBus *events.EventBusImpl,
) SessionManager {
	sm := &SessionManagerImpl{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
		cardRepo:   cardRepo,
		boardRepo:  boardRepo,
		hub:        hub,
		logger:     logger.Get(),
	}

	// NOTE: PlayerJoinedEvent subscription removed - connection handler now controls broadcast timing
	// This prevents race condition where broadcast happens before player setup is complete

	return sm
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

	// Get updated game state from NEW repository
	newGame, err := sm.gameRepo.GetByID(ctx, gameID)
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

	// Convert NEW game type to OLD model type for DTO compatibility
	game := gameToModel(newGame)

	// Get board from NEW board repository and attach to game
	sessionBoard, err := sm.boardRepo.GetByGameID(ctx, gameID)
	if err != nil {
		// Board may not exist for games created before board migration
		// Log warning but continue with empty board
		log.Warn("⚠️  Failed to get board for game, using empty board", zap.Error(err))
	} else {
		// Convert session board to model board and attach to game
		game.Board = boardToModel(sessionBoard)
		log.Debug("🗺️  Fetched board for game", zap.Int("tile_count", len(game.Board.Tiles)))
	}

	// Get all players for personalized game states from NEW repository
	newPlayers, err := sm.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players for broadcast", zap.Error(err))
		return err
	}

	// Convert NEW player types to OLD model types for DTO compatibility
	players := playersToModel(newPlayers)

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

	newResolvedCards, err := sm.cardRepo.ListCardsByIdMap(ctx, allCardIds)
	if err != nil {
		log.Error("Failed to resolve card data for broadcast", zap.Error(err))
		return err
	}

	// Convert NEW card types to OLD model types for DTO compatibility
	resolvedCards := cardsToModel(newResolvedCards)

	log.Debug("Resolved cards for broadcast",
		zap.Int("total_card_ids", len(allCardIds)),
		zap.Int("resolved_cards", len(resolvedCards)))

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

		personalizedGameDTO := dto.ToGameDto(game, players, player.ID, resolvedCards, paymentConstants)

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

// ========== Type Converters: NEW repositories → OLD model types ==========
// These converters bridge the gap between NEW repository types (internal/session/game/*)
// and OLD model types (internal/model/*) during the phased migration.

// gameToModel converts a NEW game.Game pointer to an OLD model.Game value
func gameToModel(g *game.Game) model.Game {
	return model.Game{
		ID:               g.ID,
		CreatedAt:        g.CreatedAt,
		UpdatedAt:        g.UpdatedAt,
		Status:           model.GameStatus(g.Status),
		Settings:         gameSettingsToModel(g.Settings),
		PlayerIDs:        g.PlayerIDs,
		HostPlayerID:     g.HostPlayerID,
		CurrentPhase:     model.GamePhase(g.CurrentPhase),
		GlobalParameters: g.GlobalParameters, // Same type in both systems
		ViewingPlayerID:  g.ViewingPlayerID,
		CurrentTurn:      g.CurrentTurn,
		Generation:       g.Generation,
		Board:            g.Board, // Same type in both systems
	}
}

// gameSettingsToModel converts game settings from NEW to OLD type
func gameSettingsToModel(s game.GameSettings) model.GameSettings {
	return model.GameSettings{
		MaxPlayers:      s.MaxPlayers,
		Temperature:     s.Temperature,
		Oxygen:          s.Oxygen,
		Oceans:          s.Oceans,
		DevelopmentMode: s.DevelopmentMode,
		CardPacks:       s.CardPacks,
	}
}

// playersToModel converts a slice of NEW player.Player pointers to a slice of OLD model.Player values
func playersToModel(players []*player.Player) []model.Player {
	result := make([]model.Player, len(players))
	for i, p := range players {
		result[i] = playerToModel(p)
	}
	return result
}

// playerToModel converts a NEW player.Player pointer to an OLD model.Player value
// Fields that don't exist in NEW player type are initialized with zero/empty values
func playerToModel(p *player.Player) model.Player {
	var selectStartingCards *model.SelectStartingCardsPhase

	// Convert SelectStartingCardsPhase if it exists
	if p.SelectStartingCardsPhase != nil {
		selectStartingCards = &model.SelectStartingCardsPhase{
			AvailableCards:        p.SelectStartingCardsPhase.AvailableCards,
			AvailableCorporations: p.SelectStartingCardsPhase.AvailableCorporations,
		}
	}

	// Convert PendingCardSelection if it exists
	var pendingCardSelection *model.PendingCardSelection
	if p.PendingCardSelection != nil {
		pendingCardSelection = &model.PendingCardSelection{
			Source:         p.PendingCardSelection.Source,
			AvailableCards: p.PendingCardSelection.AvailableCards,
			CardCosts:      p.PendingCardSelection.CardCosts,
			CardRewards:    p.PendingCardSelection.CardRewards,
			MinCards:       p.PendingCardSelection.MinCards,
			MaxCards:       p.PendingCardSelection.MaxCards,
		}
	}

	return model.Player{
		ID:                        p.ID,
		Name:                      p.Name,
		Corporation:               p.Corporation, // Now exists in NEW player type
		Cards:                     p.Cards,
		Resources:                 p.Resources,
		Production:                p.Production,
		TerraformRating:           p.TerraformRating,
		PlayedCards:               p.PlayedCards,      // Now exists in NEW player type
		Passed:                    p.Passed,           // Now exists in NEW player type
		AvailableActions:          p.AvailableActions, // Now exists in NEW player type
		VictoryPoints:             p.VictoryPoints,    // Now exists in NEW player type
		IsConnected:               p.IsConnected,
		Effects:                   p.Effects,         // Now exists in NEW player type
		Actions:                   p.Actions,         // Now exists in NEW player type
		ProductionPhase:           p.ProductionPhase, // Now exists in NEW player type
		SelectStartingCardsPhase:  selectStartingCards,
		PendingTileSelection:      p.PendingTileSelection,      // Now exists in NEW player type
		PendingTileSelectionQueue: p.PendingTileSelectionQueue, // Now exists in NEW player type
		PendingCardSelection:      pendingCardSelection,        // Now converted from NEW player type
		PendingCardDrawSelection:  p.PendingCardDrawSelection,  // Now exists in NEW player type
		ForcedFirstAction:         p.ForcedFirstAction,         // Now exists in NEW player type
		ResourceStorage:           p.ResourceStorage,           // Now exists in NEW player type
		PaymentSubstitutes:        p.PaymentSubstitutes,        // Now exists in NEW player type
		RequirementModifiers:      p.RequirementModifiers,      // Now exists in NEW player type
	}
}

// cardsToModel converts a map of NEW card.Card values to a map of OLD model.Card values
// The session layer Card now contains complete card data, so we copy all fields
func cardsToModel(cards map[string]card.Card) map[string]model.Card {
	result := make(map[string]model.Card, len(cards))
	for id, c := range cards {
		// Convert session card (with all data) back to model.Card
		result[id] = model.Card{
			ID:                 c.ID,
			Name:               c.Name,
			Type:               model.CardType(c.Type),
			Cost:               c.Cost,
			Description:        c.Description,
			Pack:               c.Pack,
			Tags:               c.Tags,
			Requirements:       c.Requirements,
			Behaviors:          c.Behaviors,
			ResourceStorage:    c.ResourceStorage,
			VPConditions:       c.VPConditions,
			StartingResources:  c.StartingResources,
			StartingProduction: c.StartingProduction,
		}
	}
	return result
}

// boardToModel converts a NEW board.Board pointer to an OLD model.Board value
func boardToModel(b *board.Board) model.Board {
	if b == nil {
		return model.Board{Tiles: []model.Tile{}}
	}

	tiles := make([]model.Tile, len(b.Tiles))
	for i, tile := range b.Tiles {
		// Convert bonuses
		bonuses := make([]model.TileBonus, len(tile.Bonuses))
		for j, bonus := range tile.Bonuses {
			bonuses[j] = model.TileBonus{
				Type:   model.ResourceType(bonus.Type),
				Amount: bonus.Amount,
			}
		}

		// Convert occupant if exists
		var occupant *model.TileOccupant
		if tile.OccupiedBy != nil {
			tags := make([]string, len(tile.OccupiedBy.Tags))
			copy(tags, tile.OccupiedBy.Tags)
			occupant = &model.TileOccupant{
				Type: model.ResourceType(tile.OccupiedBy.Type),
				Tags: tags,
			}
		}

		// Convert tile
		tiles[i] = model.Tile{
			Coordinates: model.HexPosition{
				Q: tile.Coordinates.Q,
				R: tile.Coordinates.R,
				S: tile.Coordinates.S,
			},
			Tags:        tile.Tags,
			Type:        model.ResourceType(tile.Type),
			Location:    model.TileLocation(tile.Location),
			DisplayName: tile.DisplayName,
			Bonuses:     bonuses,
			OccupiedBy:  occupant,
			OwnerID:     tile.OwnerID,
		}
	}

	return model.Board{Tiles: tiles}
}
