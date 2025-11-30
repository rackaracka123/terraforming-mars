package action

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/deck"
)

// CreateGameAction handles the business logic for creating new games
// MIGRATION: Uses new architecture (GameRepository only, no board repository)
type CreateGameAction struct {
	gameRepo     game.GameRepository
	eventBus     *events.EventBusImpl
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewCreateGameAction creates a new create game action
func NewCreateGameAction(
	gameRepo game.GameRepository,
	eventBus *events.EventBusImpl,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *CreateGameAction {
	return &CreateGameAction{
		gameRepo:     gameRepo,
		eventBus:     eventBus,
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// Execute performs the create game action
func (a *CreateGameAction) Execute(
	ctx context.Context,
	settings game.GameSettings,
) (*game.Game, error) {
	log := a.logger.With(
		zap.Int("max_players", settings.MaxPlayers),
		zap.Strings("card_packs", settings.CardPacks),
	)
	log.Info("🎮 Creating new game")

	// 1. Generate game ID
	gameID := uuid.New().String()

	// 2. Apply default settings
	if settings.MaxPlayers == 0 {
		settings.MaxPlayers = game.DefaultMaxPlayers
	}
	if len(settings.CardPacks) == 0 {
		settings.CardPacks = game.DefaultCardPacks()
	}

	// 3. Create game entity
	// Note: hostPlayerID is empty initially, will be set when first player joins
	// Board is automatically created by NewGame
	newGame := game.NewGame(gameID, "", settings, a.eventBus)

	// 4. Initialize deck with cards from selected packs
	projectCardIDs, corpIDs, preludeIDs := a.getCardIDsByPacks(settings.CardPacks)
	gameDeck := deck.NewDeck(gameID, projectCardIDs, corpIDs, preludeIDs)
	newGame.SetDeck(gameDeck)
	log.Info("✅ Deck initialized",
		zap.Int("project_cards", len(projectCardIDs)),
		zap.Int("corporations", len(corpIDs)),
		zap.Int("preludes", len(preludeIDs)),
		zap.Strings("first_5_corps", getFirst5(corpIDs)))

	// 5. Store game in repository
	err := a.gameRepo.Create(ctx, newGame)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		return nil, err
	}

	log.Info("✅ Game created successfully with board and deck", zap.String("game_id", gameID))
	return newGame, nil
}

// getFirst5 returns up to the first 5 elements of a slice (for logging)
func getFirst5(ids []string) []string {
	if len(ids) <= 5 {
		return ids
	}
	return ids[:5]
}

// getCardIDsByPacks retrieves card IDs filtered by pack and separated by type
func (a *CreateGameAction) getCardIDsByPacks(packs []string) (projectCards, corps, preludes []string) {
	allCards := a.cardRegistry.GetAll()

	log := a.logger.With(
		zap.Strings("requested_packs", packs),
		zap.Int("total_cards", len(allCards)),
	)
	log.Debug("🔍 Filtering cards by packs")

	// Create a map for quick pack lookup
	packMap := make(map[string]bool, len(packs))
	for _, pack := range packs {
		packMap[pack] = true
	}

	projectCards = []string{}
	corps = []string{}
	preludes = []string{}

	for _, card := range allCards {
		// Skip cards not in selected packs
		if !packMap[card.Pack] {
			continue
		}

		switch card.Type {
		case gamecards.CardTypeCorporation:
			corps = append(corps, card.ID)
			log.Debug("   Found corporation", zap.String("id", card.ID), zap.String("name", card.Name))
		case gamecards.CardTypePrelude:
			preludes = append(preludes, card.ID)
		default:
			// All other card types are project cards (Automated, Active, Event)
			projectCards = append(projectCards, card.ID)
		}
	}

	log.Info("📊 Card filtering complete",
		zap.Int("project_cards", len(projectCards)),
		zap.Int("corporations", len(corps)),
		zap.Int("preludes", len(preludes)))

	return projectCards, corps, preludes
}
