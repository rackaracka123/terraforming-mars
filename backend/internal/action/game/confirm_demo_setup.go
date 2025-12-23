package game

import (
	"context"
	"fmt"
	"math/rand"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/dto"
	internalgame "terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
)

// ConfirmDemoSetupAction handles a player confirming their demo setup configuration
type ConfirmDemoSetupAction struct {
	gameRepo     internalgame.GameRepository
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewConfirmDemoSetupAction creates a new confirm demo setup action
func NewConfirmDemoSetupAction(
	gameRepo internalgame.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *ConfirmDemoSetupAction {
	return &ConfirmDemoSetupAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// Execute performs the confirm demo setup action
func (a *ConfirmDemoSetupAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
	request *dto.ConfirmDemoSetupRequest,
) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "confirm_demo_setup"),
	)
	log.Info("🎮 Player confirming demo setup")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. Validate game is in DemoSetup phase
	if g.CurrentPhase() != internalgame.GamePhaseDemoSetup {
		log.Warn("Game is not in demo setup phase", zap.String("phase", string(g.CurrentPhase())))
		return fmt.Errorf("game is not in demo setup phase: %s", g.CurrentPhase())
	}

	// 3. Get the player
	p, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. Set corporation - either specified or random
	if request.CorporationID != nil && *request.CorporationID != "" {
		p.SetCorporationID(*request.CorporationID)
		log.Info("✅ Set corporation", zap.String("corporation_id", *request.CorporationID))
	} else {
		// Select random corporation by filtering all cards for corporation type
		allCards := a.cardRegistry.GetAll()
		var corporations []gamecards.Card
		for _, card := range allCards {
			if card.Type == gamecards.CardTypeCorporation {
				corporations = append(corporations, card)
			}
		}
		if len(corporations) > 0 {
			randomIndex := rand.Intn(len(corporations))
			randomCorpID := corporations[randomIndex].ID
			p.SetCorporationID(randomCorpID)
			log.Info("✅ Set random corporation", zap.String("corporation_id", randomCorpID))
		}
	}

	// 5. Add cards to hand with proper PlayerCard caching (using shared helper)
	if len(request.CardIDs) > 0 {
		action.AddCardsToPlayerHand(request.CardIDs, p, g, a.cardRegistry, log)
		log.Info("✅ Added cards to hand", zap.Int("card_count", len(request.CardIDs)))
	}

	// 6. Set resources
	resources := shared.Resources{
		Credits:  request.Resources.Credits,
		Steel:    request.Resources.Steel,
		Titanium: request.Resources.Titanium,
		Plants:   request.Resources.Plants,
		Energy:   request.Resources.Energy,
		Heat:     request.Resources.Heat,
	}
	p.Resources().Set(resources)
	log.Info("✅ Set resources", zap.Any("resources", resources))

	// 7. Set production
	production := shared.Production{
		Credits:  request.Production.Credits,
		Steel:    request.Production.Steel,
		Titanium: request.Production.Titanium,
		Plants:   request.Production.Plants,
		Energy:   request.Production.Energy,
		Heat:     request.Production.Heat,
	}
	p.Resources().SetProduction(production)
	log.Info("✅ Set production", zap.Any("production", production))

	// 8. Set terraform rating
	p.Resources().SetTerraformRating(request.TerraformRating)
	log.Info("✅ Set terraform rating", zap.Int("rating", request.TerraformRating))

	// 9. If host, set global parameters and generation
	isHost := g.HostPlayerID() == playerID
	if isHost && request.GlobalParameters != nil {
		gp := g.GlobalParameters()
		if err := gp.SetTemperature(ctx, request.GlobalParameters.Temperature); err != nil {
			log.Error("Failed to set temperature", zap.Error(err))
			return fmt.Errorf("failed to set temperature: %w", err)
		}
		if err := gp.SetOxygen(ctx, request.GlobalParameters.Oxygen); err != nil {
			log.Error("Failed to set oxygen", zap.Error(err))
			return fmt.Errorf("failed to set oxygen: %w", err)
		}
		if err := gp.SetOceans(ctx, request.GlobalParameters.Oceans); err != nil {
			log.Error("Failed to set oceans", zap.Error(err))
			return fmt.Errorf("failed to set oceans: %w", err)
		}
		log.Info("✅ Set global parameters",
			zap.Int("temperature", request.GlobalParameters.Temperature),
			zap.Int("oxygen", request.GlobalParameters.Oxygen),
			zap.Int("oceans", request.GlobalParameters.Oceans))
	}

	if isHost && request.Generation != nil {
		if err := g.SetGeneration(ctx, *request.Generation); err != nil {
			log.Error("Failed to set generation", zap.Error(err))
			return fmt.Errorf("failed to set generation: %w", err)
		}
		log.Info("✅ Set generation", zap.Int("generation", *request.Generation))
	}

	// 10. Mark player as having confirmed demo setup
	p.SetDemoSetupConfirmed(true)
	log.Info("✅ Player confirmed demo setup")

	// 11. Check if all players have confirmed
	allConfirmed := true
	for _, pl := range g.GetAllPlayers() {
		if !pl.DemoSetupConfirmed() {
			allConfirmed = false
			break
		}
	}

	// 12. If all players confirmed, transition to Action phase
	if allConfirmed {
		if err := g.UpdatePhase(ctx, internalgame.GamePhaseAction); err != nil {
			log.Error("Failed to update game phase", zap.Error(err))
			return fmt.Errorf("failed to update game phase: %w", err)
		}
		log.Info("🎉 All players confirmed, transitioning to Action phase")

		// Initialize milestones and awards for all players (event-driven state caching)
		allPlayers := g.GetAllPlayers()
		for _, pl := range allPlayers {
			action.InitializePlayerMilestones(pl, g, a.cardRegistry)
			action.InitializePlayerAwards(pl, g)
		}
		log.Info("✅ Milestones and awards initialized for all players")

		// Set first player turn with appropriate action count
		if len(allPlayers) > 0 {
			firstPlayerID := allPlayers[0].ID()
			availableActions := 2
			if len(allPlayers) == 1 {
				availableActions = -1 // Unlimited for solo mode
			}
			if err := g.SetCurrentTurn(ctx, firstPlayerID, availableActions); err != nil {
				log.Error("Failed to set current turn", zap.Error(err))
				return fmt.Errorf("failed to set current turn: %w", err)
			}
			log.Info("✅ Set first player turn with actions",
				zap.String("player_id", firstPlayerID),
				zap.Int("available_actions", availableActions))
		}
	}

	return nil
}
