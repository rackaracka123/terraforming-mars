package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/game/player"

	"go.uber.org/zap"
)

// StartGameAction handles the business logic for starting games
type StartGameAction struct {
	gameRepo   game.Repository
	playerRepo player.Repository
	cardRepo   card.Repository
	sessionMgr session.SessionManager
	logger     *zap.Logger
}

// NewStartGameAction creates a new start game action
func NewStartGameAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	cardRepo card.Repository,
	sessionMgr session.SessionManager,
) *StartGameAction {
	return &StartGameAction{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
		cardRepo:   cardRepo,
		sessionMgr: sessionMgr,
		logger:     logger.Get(),
	}
}

// Execute performs the start game action
func (a *StartGameAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
	)
	log.Info("🎮 Starting game")

	// 1. Validate business rules: player must be host
	g, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return fmt.Errorf("game not found: %w", err)
	}

	if g.HostPlayerID != playerID {
		log.Error("Non-host attempted to start game")
		return fmt.Errorf("only the host can start the game")
	}

	// 2. Validate game status
	if g.Status != game.GameStatusLobby {
		log.Error("Game not in lobby status", zap.String("status", string(g.Status)))
		return fmt.Errorf("game not in lobby status")
	}

	// 3. Get players
	players, err := a.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players", zap.Error(err))
		return fmt.Errorf("failed to get players: %w", err)
	}

	if len(players) < 1 {
		log.Error("No players in game")
		return fmt.Errorf("cannot start game with no players")
	}

	// 4. Update game status to Active (event-driven)
	err = a.gameRepo.UpdateStatus(ctx, gameID, game.GameStatusActive)
	if err != nil {
		log.Error("Failed to update game status", zap.Error(err))
		return fmt.Errorf("failed to update game status: %w", err)
	}

	log.Info("✅ Game status updated to Active")

	// 5. Update game phase to StartingCardSelection (event-driven)
	err = a.gameRepo.UpdatePhase(ctx, gameID, game.GamePhaseStartingCardSelection)
	if err != nil {
		log.Error("Failed to update game phase", zap.Error(err))
		return fmt.Errorf("failed to update game phase: %w", err)
	}

	log.Info("✅ Game phase updated to StartingCardSelection")

	// 6. Set first player's turn
	if len(players) > 0 {
		firstPlayerID := players[0].ID
		err = a.gameRepo.SetCurrentTurn(ctx, gameID, &firstPlayerID)
		if err != nil {
			log.Error("Failed to set current turn", zap.Error(err))
			return fmt.Errorf("failed to set current turn: %w", err)
		}
		log.Info("✅ Set initial turn", zap.String("first_player_id", firstPlayerID))
	}

	// 7. Distribute starting cards to all players
	err = a.distributeStartingCards(ctx, gameID, players)
	if err != nil {
		log.Error("Failed to distribute starting cards", zap.Error(err))
		return fmt.Errorf("failed to distribute starting cards: %w", err)
	}

	log.Info("✅ Starting cards distributed to all players")

	// 8. Broadcast state via session manager
	err = a.sessionMgr.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Non-fatal, continue
	}

	log.Info("🎉 Game started successfully")
	return nil
}

// distributeStartingCards gives each player 10 project cards and 2 corporations
func (a *StartGameAction) distributeStartingCards(ctx context.Context, gameID string, players []*player.Player) error {
	log := a.logger.With(zap.String("game_id", gameID))
	log.Debug("Distributing starting cards to players", zap.Int("player_count", len(players)))

	for _, p := range players {
		// Draw 10 project cards
		projectCards, err := a.cardRepo.DrawProjectCards(ctx, 10)
		if err != nil {
			return fmt.Errorf("failed to draw project cards for player %s: %w", p.ID, err)
		}

		// Draw 2 corporation cards
		corporations, err := a.cardRepo.DrawCorporations(ctx, 2)
		if err != nil {
			return fmt.Errorf("failed to draw corporations for player %s: %w", p.ID, err)
		}

		// Extract card IDs
		projectCardIDs := make([]string, len(projectCards))
		for i, c := range projectCards {
			projectCardIDs[i] = c.ID
		}

		corporationIDs := make([]string, len(corporations))
		for i, c := range corporations {
			corporationIDs[i] = c.ID
		}

		// Set starting cards selection phase for player
		err = a.playerRepo.SetStartingCardsSelection(ctx, gameID, p.ID, projectCardIDs, corporationIDs)
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
