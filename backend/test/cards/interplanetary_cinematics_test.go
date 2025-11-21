package cards_test

import (
	"context"
	"fmt"
	"testing"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/session/card"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterplanetaryCinematicsTriggersOnEventCards(t *testing.T) {
	ctx := context.Background()

	// Setup event bus and repositories
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	// Setup CardEffectSubscriber
	effectSubscriber := card.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo, cardRepo)

	// Create game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	// Create player with Interplanetary Cinematics
	playerID := "player-1"
	player := model.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        model.Resources{Credits: 50},
		Production:       model.Production{},
		TerraformRating:  20,
		AvailableActions: 2,
	}
	err = playerRepo.Create(ctx, game.ID, player)
	require.NoError(t, err)

	// Create Interplanetary Cinematics card with card-played trigger
	interplanetaryCinematicsCard := &model.Card{
		ID:   "B04",
		Name: "Interplanetary Cinematics",
		Type: model.CardTypeCorporation,
		Behaviors: []model.CardBehavior{
			{
				// First behavior: starting resources
				Triggers: []model.Trigger{{
					Type: model.ResourceTriggerAuto,
				}},
				Outputs: []model.ResourceCondition{
					{Type: model.ResourceCredits, Amount: 30, Target: model.TargetSelfPlayer},
					{Type: model.ResourceSteel, Amount: 20, Target: model.TargetSelfPlayer},
				},
			},
			{
				// Second behavior: gain 2 credits when event card played
				Triggers: []model.Trigger{{
					Type: model.ResourceTriggerAuto,
					Condition: &model.ResourceTriggerCondition{
						Type:              model.TriggerCardPlayed,
						AffectedCardTypes: []model.CardType{model.CardTypeEvent},
					},
				}},
				Outputs: []model.ResourceCondition{
					{Type: model.ResourceCredits, Amount: 2, Target: model.TargetSelfPlayer},
				},
			},
		},
	}

	// Subscribe passive effects for Interplanetary Cinematics
	// This simulates the corp being selected and its passive effect being registered
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, playerID, "B04", interplanetaryCinematicsCard)
	require.NoError(t, err)

	// Get initial credits
	player, err = playerRepo.GetByID(ctx, game.ID, playerID)
	require.NoError(t, err)
	initialCredits := player.Resources.Credits

	// Simulate playing an event card (should trigger +2 credits)
	// We need to publish a CardPlayedEvent with type "event"
	cardPlayedEvent := repository.CardPlayedEvent{
		GameID:   game.ID,
		PlayerID: playerID,
		CardID:   "event-card-1",
		CardName: "Test Event Card",
		CardType: string(model.CardTypeEvent),
	}
	events.Publish(eventBus, cardPlayedEvent)

	// Verify player gained 2 credits
	player, err = playerRepo.GetByID(ctx, game.ID, playerID)
	require.NoError(t, err)
	assert.Equal(t, initialCredits+2, player.Resources.Credits, "Player should gain 2 credits when playing an event card")

	// Simulate playing a non-event card (should NOT trigger +2 credits)
	cardPlayedEvent2 := repository.CardPlayedEvent{
		GameID:   game.ID,
		PlayerID: playerID,
		CardID:   "automated-card-1",
		CardName: "Test Automated Card",
		CardType: string(model.CardTypeAutomated),
	}
	events.Publish(eventBus, cardPlayedEvent2)

	// Verify player did NOT gain credits
	player, err = playerRepo.GetByID(ctx, game.ID, playerID)
	require.NoError(t, err)
	assert.Equal(t, initialCredits+2, player.Resources.Credits, "Player should NOT gain credits when playing a non-event card")
}

func TestInterplanetaryCinematicsMultipleEventCards(t *testing.T) {
	ctx := context.Background()

	// Setup event bus and repositories
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	// Setup CardEffectSubscriber
	effectSubscriber := card.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo, cardRepo)

	// Create game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	// Create player with Interplanetary Cinematics
	playerID := "player-1"
	player := model.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        model.Resources{Credits: 50},
		Production:       model.Production{},
		TerraformRating:  20,
		AvailableActions: 2,
	}
	err = playerRepo.Create(ctx, game.ID, player)
	require.NoError(t, err)

	// Create Interplanetary Cinematics card with card-played trigger
	interplanetaryCinematicsCard := &model.Card{
		ID:   "B04",
		Name: "Interplanetary Cinematics",
		Type: model.CardTypeCorporation,
		Behaviors: []model.CardBehavior{
			{
				// Gain 2 credits when event card played
				Triggers: []model.Trigger{{
					Type: model.ResourceTriggerAuto,
					Condition: &model.ResourceTriggerCondition{
						Type:              model.TriggerCardPlayed,
						AffectedCardTypes: []model.CardType{model.CardTypeEvent},
					},
				}},
				Outputs: []model.ResourceCondition{
					{Type: model.ResourceCredits, Amount: 2, Target: model.TargetSelfPlayer},
				},
			},
		},
	}

	// Subscribe passive effects
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, playerID, "B04", interplanetaryCinematicsCard)
	require.NoError(t, err)

	// Get initial credits
	player, err = playerRepo.GetByID(ctx, game.ID, playerID)
	require.NoError(t, err)
	initialCredits := player.Resources.Credits

	// Play 3 event cards
	for i := 0; i < 3; i++ {
		cardPlayedEvent := repository.CardPlayedEvent{
			GameID:   game.ID,
			PlayerID: playerID,
			CardID:   fmt.Sprintf("event-card-%d", i+1),
			CardName: "Test Event Card",
			CardType: string(model.CardTypeEvent),
		}
		events.Publish(eventBus, cardPlayedEvent)
	}

	// Verify player gained 6 credits (2 per event card)
	player, err = playerRepo.GetByID(ctx, game.ID, playerID)
	require.NoError(t, err)
	assert.Equal(t, initialCredits+6, player.Resources.Credits, "Player should gain 6 credits when playing 3 event cards")
}
