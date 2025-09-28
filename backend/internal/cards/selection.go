package cards

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// SelectionManager handles all card selection operations
type SelectionManager struct {
	gameRepo     repository.GameRepository
	playerRepo   repository.PlayerRepository
	cardRepo     repository.CardRepository
	cardDeckRepo repository.CardDeckRepository
}

// NewSelectionManager creates a new card selection manager
func NewSelectionManager(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, cardRepo repository.CardRepository, cardDeckRepo repository.CardDeckRepository) *SelectionManager {
	return &SelectionManager{
		gameRepo:     gameRepo,
		playerRepo:   playerRepo,
		cardRepo:     cardRepo,
		cardDeckRepo: cardDeckRepo,
	}
}

// SelectStartingCards handles the starting card selection for a player (stores as pending, doesn't commit)
func (s *SelectionManager) SelectStartingCards(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Processing starting card selection (storing as pending)", zap.Strings("card_ids", cardIDs))

	// Validate the selection
	if err := s.ValidateStartingCardSelection(ctx, gameID, playerID, cardIDs); err != nil {
		log.Error("Starting card selection validation failed", zap.Error(err))
		return fmt.Errorf("invalid card selection: %w", err)
	}

	// Get current player
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Calculate cost (3 MC per card)
	cost := len(cardIDs) * 3

	// Check if player can afford the selection
	if player.Resources.Credits < cost {
		return fmt.Errorf("insufficient credits: need %d, have %d", cost, player.Resources.Credits)
	}

	// Update player resources immediately (deduct credits)
	updatedResources := player.Resources
	updatedResources.Credits -= cost
	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, updatedResources); err != nil {
		log.Error("Failed to update player resources", zap.Error(err))
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	// Add selected cards to player's hand immediately using granular updates
	for _, cardID := range cardIDs {
		if err := s.playerRepo.AddCard(ctx, gameID, playerID, cardID); err != nil {
			log.Error("Failed to add card to player hand", zap.String("card_id", cardID), zap.Error(err))
			return fmt.Errorf("failed to add card %s: %w", cardID, err)
		}
	}

	// Clear the starting selection (this hides the modal)
	if err := s.playerRepo.SetStartingCardsSelectionComplete(ctx, gameID, playerID); err != nil {
		log.Error("Failed to clear starting selection", zap.Error(err))
		return fmt.Errorf("failed to clear starting selection: %w", err)
	}

	log.Info("Player starting card selection completed",
		zap.Strings("selected_cards", cardIDs),
		zap.Int("cost", cost))

	return nil
}

// SelectProductionCards handles the card selection during production phase
func (s *SelectionManager) SelectProductionCards(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Processing production card selection", zap.Strings("card_ids", cardIDs))

	// Get current player
	_, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Add selected cards to player's hand using granular updates
	for _, cardID := range cardIDs {
		if err := s.playerRepo.AddCard(ctx, gameID, playerID, cardID); err != nil {
			log.Error("Failed to add card to player hand", zap.String("card_id", cardID), zap.Error(err))
			return fmt.Errorf("failed to add card %s: %w", cardID, err)
		}
	}

	// Mark player as ready for production phase
	err = s.playerRepo.SetProductionCardsSelectionComplete(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to mark production card selection complete", zap.Error(err))
		return fmt.Errorf("failed to mark production card selection complete: %w", err)
	}

	log.Info("Player production card selection completed",
		zap.Strings("selected_cards", cardIDs))

	return nil
}

// ValidateStartingCardSelection validates a player's starting card selection
func (s *SelectionManager) ValidateStartingCardSelection(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player to check starting card selection state
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Check if player already has cards in hand (indicating they already completed selection)
	if len(player.Cards) > 0 {
		log.Debug("Player has cards in hand, selection already completed", zap.Int("cards_in_hand", len(player.Cards)))
		return fmt.Errorf("starting card selection already completed - you have %d cards in your hand", len(player.Cards))
	}

	if player.SelectStartingCardsPhase == nil {
		return fmt.Errorf("starting card selection phase not initialized for player")
	}

	// Check if player has starting cards available for selection
	if len(player.SelectStartingCardsPhase.AvailableCards) == 0 {
		log.Debug("Player has no starting cards available",
			zap.Int("cards_in_hand", len(player.Cards)),
			zap.Bool("has_starting_selection", len(player.SelectStartingCardsPhase.AvailableCards) > 0))
		return fmt.Errorf("no starting cards available for selection - selection phase may have ended")
	}

	playerOptions := player.SelectStartingCardsPhase.AvailableCards

	// Check maximum cards (10 is the maximum starting cards dealt)
	if len(cardIDs) > 10 {
		return fmt.Errorf("cannot select more than 10 cards, got %d", len(cardIDs))
	}

	// Validate card IDs exist in the starting cards pool first
	allStartingCards, err := s.cardRepo.GetStartingCardPool(ctx)
	if err != nil {
		return fmt.Errorf("failed to get starting card pool: %w", err)
	}
	cardMap := make(map[string]bool)
	for _, card := range allStartingCards {
		cardMap[card.ID] = true
	}

	for _, cardID := range cardIDs {
		if !cardMap[cardID] {
			return fmt.Errorf("invalid card ID: %s", cardID)
		}
	}

	// Then validate selected cards are in player's options
	optionsMap := make(map[string]bool)
	for _, cardId := range playerOptions {
		optionsMap[cardId] = true
	}

	for _, cardID := range cardIDs {
		if !optionsMap[cardID] {
			return fmt.Errorf("card %s is not in player's available options", cardID)
		}
	}

	return nil
}

// IsAllPlayersCardSelectionComplete checks if all players in the game have completed card selection
func (s *SelectionManager) IsAllPlayersCardSelectionComplete(ctx context.Context, gameID string) bool {
	log := logger.WithGameContext(gameID, "")

	// Get current game state to determine which selection type to check
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for selection completion check", zap.Error(err))
		return false
	}

	// Get all players in the game
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players for selection completion check", zap.Error(err))
		return false
	}

	// If no players exist, selection is not complete
	if len(players) == 0 {
		return false
	}

	// Check completion based on current game phase
	switch game.CurrentPhase {
	case model.GamePhaseStartingCardSelection:
		return s.checkStartingCardSelectionComplete(players, log)
	case model.GamePhaseProductionAndCardDraw:
		return s.checkProductionCardSelectionComplete(players)
	default:
		// For other phases, assume no card selection is needed
		return true
	}
}

// checkStartingCardSelectionComplete checks if all players have completed starting card selection using the flag
func (s *SelectionManager) checkStartingCardSelectionComplete(players []model.Player, log *zap.Logger) bool {
	playersCompleted := 0

	// Check the HasSelectedStartingCards flag for each player
	for _, player := range players {
		if player.SelectStartingCardsPhase == nil {
			log.Debug("Player has no starting card selection phase initialized", zap.String("player_id", player.ID))
			continue
		}

		if player.SelectStartingCardsPhase.SelectionComplete {
			playersCompleted++
			log.Debug("Player has completed starting card selection", zap.String("player_id", player.ID))
		} else if len(player.SelectStartingCardsPhase.AvailableCards) > 0 {
			log.Debug("Player has starting selection but hasn't completed yet", zap.String("player_id", player.ID))
		} else {
			log.Debug("Player has no starting selection", zap.String("player_id", player.ID))
		}
	}

	// All players must have completed their starting card selection
	allComplete := playersCompleted == len(players)
	log.Debug("Starting card selection completion check",
		zap.Int("total_players", len(players)),
		zap.Int("players_completed", playersCompleted),
		zap.Bool("all_complete", allComplete))

	return allComplete
}

// checkProductionCardSelectionComplete checks production card selection completion
func (s *SelectionManager) checkProductionCardSelectionComplete(players []model.Player) bool {
	// Track selection states
	playersWithIncompleteSelection := 0
	playersWithSelectionData := 0

	// Check each player's selection status
	for _, player := range players {
		if player.ProductionPhase != nil {
			playersWithSelectionData++
			// If any player has selection data but hasn't completed, selection is not done
			if !player.ProductionPhase.SelectionComplete {
				playersWithIncompleteSelection++
			}
		}
	}

	// If any player has incomplete selection, overall selection is not complete
	if playersWithIncompleteSelection > 0 {
		return false
	}

	// If no players have selection data at all, selection is not complete
	// This handles the case where the selection phase hasn't been initiated yet
	if playersWithSelectionData == 0 {
		return false
	}

	// If we reach here, all players with selection data have completed selection
	return true
}

// ClearGameSelectionData clears temporary selection data for a game (called after selection phase completes)
func (s *SelectionManager) ClearGameSelectionData(gameID string) {
	ctx := context.Background()

	// Get all players in the game and clear their selection data
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		logger.WithGameContext(gameID, "").Error("Failed to get players for clearing selection data", zap.Error(err))
		return
	}

	for _, player := range players {
		if player.SelectStartingCardsPhase != nil {
			err = s.playerRepo.UpdateSelectStartingCardsPhase(ctx, gameID, player.ID, nil)
			if err != nil {
				logger.WithGameContext(gameID, player.ID).Error("Failed to clear starting card selection data", zap.Error(err))
				return
			}
		}

		if player.ProductionPhase != nil {

			err = s.playerRepo.UpdateProductionPhase(ctx, gameID, player.ID, nil)
			if err != nil {
				logger.WithGameContext(gameID, player.ID).Error("Failed to clear production phase data", zap.Error(err))
				return
			}
		}
	}

	logger.WithGameContext(gameID, "").Debug("Cleared game selection data")
}
