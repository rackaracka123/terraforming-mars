package initialization

import (
	"fmt"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/cards/economy"
	"terraforming-mars-backend/internal/cards/plants"
	"terraforming-mars-backend/internal/cards/power"
	"terraforming-mars-backend/internal/cards/science"
	"terraforming-mars-backend/internal/cards/space"
	"terraforming-mars-backend/internal/events"
)

// RegisterAllCards registers all card handlers with the global registry
func RegisterAllCards() error {
	handlers := getAllCardHandlers()

	for _, handler := range handlers {
		if err := cards.RegisterCardHandler(handler); err != nil {
			return err
		}
	}

	return nil
}

// RegisterCardsWithRegistry registers all card handlers with a specific registry
func RegisterCardsWithRegistry(registry *cards.CardHandlerRegistry) error {
	handlers := getAllCardHandlers()

	for _, handler := range handlers {
		if err := registry.Register(handler); err != nil {
			return err
		}
	}

	return nil
}

// RegisterCardListeners registers event listeners for all cards that need them
// This automatically detects which cards implement the ListenerRegistrar interface
func RegisterCardListeners(eventBus events.EventBus) error {
	handlers := getAllCardHandlers()

	for _, handler := range handlers {
		// Check if this handler implements ListenerRegistrar
		if registrar, ok := handler.(cards.ListenerRegistrar); ok {
			if err := registrar.RegisterListeners(eventBus); err != nil {
				return fmt.Errorf("failed to register listeners for card %s: %w", handler.GetCardID(), err)
			}
		}
	}

	return nil
}

// UnregisterCardListeners cleans up event listeners for all cards
func UnregisterCardListeners(eventBus events.EventBus) error {
	handlers := getAllCardHandlers()

	for _, handler := range handlers {
		// Check if this handler implements ListenerRegistrar
		if registrar, ok := handler.(cards.ListenerRegistrar); ok {
			if err := registrar.UnregisterListeners(eventBus); err != nil {
				return fmt.Errorf("failed to unregister listeners for card %s: %w", handler.GetCardID(), err)
			}
		}
	}

	return nil
}

// getAllCardHandlers returns all card handlers for use in registration functions
func getAllCardHandlers() []cards.CardHandler {
	return []cards.CardHandler{
		// Economy cards
		economy.NewEarlySettlementHandler(),
		economy.NewInvestmentHandler(),
		economy.NewMiningOperationHandler(),
		economy.NewMiningGuildHandler(), // With listeners

		// Power cards
		power.NewPowerPlantHandler(),
		power.NewHeatGeneratorsHandler(),
		power.NewSpaceMirrorsHandler(),

		// Science cards
		science.NewResearchGrantHandler(),
		science.NewAtmosphericProcessorsHandler(), // With listeners

		// Space cards
		space.NewWaterImportHandler(),

		// Plant cards
		plants.NewNitrogenPlantsHandler(),
	}
}
