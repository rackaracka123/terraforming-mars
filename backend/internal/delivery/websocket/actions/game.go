package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"
)

// GameActions handles game lifecycle actions
type GameActions struct {
	gameService   service.GameService
	playerService service.PlayerService
	broadcaster   *core.Broadcaster
}

// NewGameActions creates a new game actions handler
func NewGameActions(gameService service.GameService, playerService service.PlayerService, broadcaster *core.Broadcaster) *GameActions {
	return &GameActions{
		gameService:   gameService,
		playerService: playerService,
		broadcaster:   broadcaster,
	}
}

// StartGame handles the start game action
func (ga *GameActions) StartGame(ctx context.Context, gameID, playerID string) error {
	return ga.gameService.StartGame(ctx, gameID, playerID)
}

// SkipAction handles the skip action
func (ga *GameActions) SkipAction(ctx context.Context, gameID, playerID string) error {
	game, err := ga.gameService.GetGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game state for phase validation: %w", err)
	}

	if game.CurrentPhase != model.GamePhaseAction {
		return fmt.Errorf("skip action not allowed in phase %s, must be in action phase", game.CurrentPhase)
	}

	result, err := ga.gameService.SkipPlayerTurn(ctx, gameID, playerID)
	if err != nil {
		return err
	}

	// Check if all players have passed - trigger production phase
	if result.AllPlayersPassed {
		gameAfterProduction, err := ga.gameService.ExecuteProductionPhase(ctx, gameID)
		if err != nil {
			return fmt.Errorf("failed to execute production phase: %w", err)
		}

		// Generate the production data for each player
		playersData := make([]dto.PlayerProductionData, len(gameAfterProduction.PlayerIDs))
		for idx, playerId := range gameAfterProduction.PlayerIDs {
			data, err := ga.ComputeProduction(ctx, gameID, playerID)
			if err != nil {
				return fmt.Errorf("failed to compute production for player %s: %w", playerId, err)
			}
			playersData[idx] = data
		}

		// Broadcast production data to all players
		ga.broadcaster.BroadcastProductionPhaseStarted(ctx, gameID, playersData)
	}

	return nil
}

// ComputeProduction handles the compute production action and returns the updated player
func (ga *GameActions) ComputeProduction(ctx context.Context, gameId, playerId string) (dto.PlayerProductionData, error) {

	playerAfterProduction, err := ga.playerService.GetPlayer(ctx, gameId, playerId)
	if err != nil {
		return dto.PlayerProductionData{}, fmt.Errorf("failed to get playerAfterProduction state for production computation: %w", err)
	}

	// Create player production data for each playerAfterProduction
	var playersData []dto.PlayerProductionData
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

	playersData = append(playersData, playerData)

	return playerData, nil
}
