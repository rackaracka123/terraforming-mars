package specialized

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core/broadcast"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"
)

// LifecycleActions handles game lifecycle actions (start, skip, production)
type LifecycleActions struct {
	gameService   service.GameService
	playerService service.PlayerService
	broadcaster   *broadcast.Broadcaster
}

// NewLifecycleActions creates a new lifecycle actions handler
func NewLifecycleActions(gameService service.GameService, playerService service.PlayerService, broadcaster *broadcast.Broadcaster) *LifecycleActions {
	return &LifecycleActions{
		gameService:   gameService,
		playerService: playerService,
		broadcaster:   broadcaster,
	}
}

// StartGame handles the start game action
func (la *LifecycleActions) StartGame(ctx context.Context, gameID, playerID string) error {
	return la.gameService.StartGame(ctx, gameID, playerID)
}

// SkipAction handles the skip action
func (la *LifecycleActions) SkipAction(ctx context.Context, gameID, playerID string) error {
	game, err := la.gameService.GetGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game state for phase validation: %w", err)
	}

	if game.CurrentPhase != model.GamePhaseAction {
		return fmt.Errorf("skip action not allowed in phase %s, must be in action phase", game.CurrentPhase)
	}

	result, err := la.gameService.SkipPlayerTurn(ctx, gameID, playerID)
	if err != nil {
		return err
	}

	// Check if all players have passed - trigger production phase
	if result.AllPlayersPassed {
		gameAfterProduction, err := la.gameService.ExecuteProductionPhase(ctx, gameID)
		if err != nil {
			return fmt.Errorf("failed to execute production phase: %w", err)
		}

		// Generate the production data for each player
		playersData := make([]dto.PlayerProductionData, len(gameAfterProduction.PlayerIDs))
		for idx, playerId := range gameAfterProduction.PlayerIDs {
			data, err := la.ComputeProduction(ctx, gameID, playerId)
			if err != nil {
				return fmt.Errorf("failed to compute production for player %s: %w", playerId, err)
			}
			playersData[idx] = data
		}

		// Broadcast production data to all players
		la.broadcaster.BroadcastProductionPhaseStarted(ctx, gameID, playersData)
	}

	return nil
}

// ComputeProduction handles the compute production action and returns the updated player
func (la *LifecycleActions) ComputeProduction(ctx context.Context, gameId, playerId string) (dto.PlayerProductionData, error) {
	playerAfterProduction, err := la.playerService.GetPlayer(ctx, gameId, playerId)
	if err != nil {
		return dto.PlayerProductionData{}, fmt.Errorf("failed to get playerAfterProduction state for production computation: %w", err)
	}

	// Calculate before resources (resources before production was applied)
	beforeResources := dto.ResourcesDto{
		Credits:  playerAfterProduction.Resources.Credits - playerAfterProduction.Production.Credits - playerAfterProduction.TerraformRating,
		Steel:    playerAfterProduction.Resources.Steel - playerAfterProduction.Production.Steel,
		Titanium: playerAfterProduction.Resources.Titanium - playerAfterProduction.Production.Titanium,
		Plants:   playerAfterProduction.Resources.Plants - playerAfterProduction.Production.Plants,
		Energy:   playerAfterProduction.Production.Energy,                                                                                // Energy before was the old energy that got converted to heat
		Heat:     playerAfterProduction.Resources.Heat - playerAfterProduction.Production.Heat - playerAfterProduction.Production.Energy, // Heat before production and energy conversion
	}

	// Energy converted is the energy that was converted to heat (equal to energy production)
	energyConverted := playerAfterProduction.Production.Energy

	// Current resources are after production
	afterResources := dto.ResourcesDto{
		Credits:  playerAfterProduction.Resources.Credits,
		Steel:    playerAfterProduction.Resources.Steel,
		Titanium: playerAfterProduction.Resources.Titanium,
		Plants:   playerAfterProduction.Resources.Plants,
		Energy:   playerAfterProduction.Resources.Energy,
		Heat:     playerAfterProduction.Resources.Heat,
	}

	production := dto.ProductionDto{
		Credits:  playerAfterProduction.Production.Credits,
		Steel:    playerAfterProduction.Production.Steel,
		Titanium: playerAfterProduction.Production.Titanium,
		Plants:   playerAfterProduction.Production.Plants,
		Energy:   playerAfterProduction.Production.Energy,
		Heat:     playerAfterProduction.Production.Heat,
	}

	// Calculate total credits income (production + terraform rating)
	creditsIncome := playerAfterProduction.Production.Credits + playerAfterProduction.TerraformRating

	playerData := dto.PlayerProductionData{
		PlayerID:        playerAfterProduction.ID,
		PlayerName:      playerAfterProduction.Name,
		BeforeResources: beforeResources,
		AfterResources:  afterResources,
		Production:      production,
		TerraformRating: playerAfterProduction.TerraformRating,
		EnergyConverted: energyConverted,
		CreditsIncome:   creditsIncome,
	}

	return playerData, nil
}