package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game/card"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/game/deck"
	"terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// convertGameSettings converts NEW game.GameSettings to OLD types.GameSettings
// Temporary helper during migration - to be removed when deckRepo uses NEW types
func convertGameSettings(newSettings game.GameSettings) types.GameSettings {
	return types.GameSettings{
		MaxPlayers:      newSettings.MaxPlayers,
		Temperature:     newSettings.Temperature,
		Oxygen:          newSettings.Oxygen,
		Oceans:          newSettings.Oceans,
		DevelopmentMode: newSettings.DevelopmentMode,
		CardPacks:       newSettings.CardPacks,
	}
}

// StartGameAction handles the business logic for starting games
type StartGameAction struct {
	BaseAction // Embed base for common dependencies and utilities
	gameRepo   game.Repository
	cardRepo   card.Repository
	deckRepo   deck.Repository
}

// NewStartGameAction creates a new start game action
func NewStartGameAction(
	gameRepo game.Repository,
	cardRepo card.Repository,
	deckRepo deck.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *StartGameAction {
	return &StartGameAction{
		BaseAction: NewBaseAction(sessionMgrFactory),
		gameRepo:   gameRepo,
		cardRepo:   cardRepo,
		deckRepo:   deckRepo,
	}
}

// Execute performs the start game action
func (a *StartGameAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID)
	log.Info("🎮 Starting game")

	// 1. Validate game is in lobby and player is host
	g, err := ValidateLobbyGame(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	if err := ValidateHostPermission(g, playerID, log); err != nil {
		return err
	}

	// 2. Get session and all players
	players := sess.GetAllPlayers()

	// No minimum player count validation - allow any number for development/testing
	log.Info("🎮 Starting game with players", zap.Int("player_count", len(players)))

	// 3. Update game status to Active
	if err := TransitionGameStatus(ctx, a.gameRepo, gameID, game.GameStatusActive, log); err != nil {
		return err
	}

	// 4. Update game phase to StartingCardSelection
	if err := TransitionGamePhase(ctx, a.gameRepo, gameID, game.GamePhaseStartingCardSelection, log); err != nil {
		return err
	}

	// 5. Set first player's turn
	if len(players) > 0 {
		firstPlayerID := players[0].ID
		if err := SetCurrentTurn(ctx, a.gameRepo, gameID, &firstPlayerID, log); err != nil {
			return err
		}
		log.Info("✅ Set initial turn", zap.String("first_player_id", firstPlayerID))
	}

	// Set unlimited actions for solo mode
	if len(players) == 1 {
		err = players[0].Action.UpdateAvailableActions(ctx, -1)
		if err != nil {
			log.Error("Failed to set unlimited actions for solo mode", zap.Error(err))
			return fmt.Errorf("failed to set unlimited actions for solo mode: %w", err)
		}
		log.Info("🎮 Solo mode detected - unlimited actions enabled")
	}

	// 6. Create game deck with shuffled cards
	err = a.deckRepo.CreateDeck(ctx, convertGameSettings(g.Settings))
	if err != nil {
		log.Error("Failed to create game deck", zap.Error(err))
		return fmt.Errorf("failed to create game deck: %w", err)
	}
	log.Info("🎴 Game deck created and shuffled")

	// 7. Distribute starting cards to all players
	if err := a.distributeStartingCards(ctx, gameID, players); err != nil {
		log.Error("Failed to distribute starting cards", zap.Error(err))
		return fmt.Errorf("failed to distribute starting cards: %w", err)
	}

	log.Info("✅ Starting cards distributed to all players")

	// 8. Broadcast state to all players
	a.BroadcastGameState(gameID, log)

	log.Info("🎉 Game started successfully")
	return nil
}

// distributeStartingCards gives each player 10 project cards and 2 corporations
func (a *StartGameAction) distributeStartingCards(ctx context.Context, gameID string, players []*player.Player) error {
	log := a.logger.With(zap.String("game_id", gameID))
	log.Debug("Distributing starting cards to players", zap.Int("player_count", len(players)))

	for _, p := range players {
		// Draw 10 project cards from game deck
		projectCardIDs, err := a.deckRepo.DrawProjectCards(ctx, 10)
		if err != nil {
			return fmt.Errorf("failed to draw project cards for player %s: %w", p.ID, err)
		}

		// Draw 2 corporation cards from game deck
		corporationIDs, err := a.deckRepo.DrawCorporations(ctx, 2)
		if err != nil {
			return fmt.Errorf("failed to draw corporations for player %s: %w", p.ID, err)
		}

		// Set starting cards selection phase for player
		err = p.Selection.SetStartingCardsSelection(ctx, projectCardIDs, corporationIDs)
		if err != nil {
			return fmt.Errorf("failed to set starting cards for player %s: %w", p.ID, err)
		}

		log.Debug("✅ Distributed cards to player",
			zap.String("player_id", p.ID),
			zap.Int("project_cards", len(projectCardIDs)),
			zap.Int("corporations", len(corporationIDs)))
	}

	return nil
}
