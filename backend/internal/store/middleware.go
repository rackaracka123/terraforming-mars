package store

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// LoggingMiddleware logs all actions and state changes
func LoggingMiddleware(next Reducer) Reducer {
	return func(state ApplicationState, action Action) (ApplicationState, error) {
		logger.Info("🔄 Applying action",
			zap.String("type", string(action.Type)),
			zap.String("game_id", action.Meta.GameID),
			zap.String("player_id", action.Meta.PlayerID))

		newState, err := next(state, action)
		if err != nil {
			logger.Error("❌ Action failed",
				zap.String("type", string(action.Type)),
				zap.Error(err))
			return state, err
		}

		logger.Debug("✅ Action applied successfully",
			zap.String("type", string(action.Type)))

		return newState, nil
	}
}

// EventMiddleware emits domain events after successful state changes
func EventMiddleware(eventBus events.EventBus) Middleware {
	return func(next Reducer) Reducer {
		return func(state ApplicationState, action Action) (ApplicationState, error) {
			// Apply the action first
			newState, err := next(state, action)
			if err != nil {
				return state, err
			}

			// Create a context for event publishing
			ctx := context.Background()

			// Emit appropriate events based on action type
			if action.Meta.GameID != "" {
				switch action.Type {
				case ActionCreateGame:
					if payload, ok := action.Payload.(CreateGamePayload); ok {
						eventBus.Publish(ctx, events.NewGameCreatedEvent(payload.GameID, payload.Settings))
					}

				case ActionStartGame:
					if gameState, exists := newState.GetGame(action.Meta.GameID); exists {
						playerCount := len(gameState.Players())
						eventBus.Publish(ctx, events.NewGameStartedEvent(action.Meta.GameID, playerCount))
					}

				case ActionJoinGame:
					if payload, ok := action.Payload.(JoinGamePayload); ok {
						eventBus.Publish(ctx, events.NewPlayerJoinedEvent(payload.GameID, payload.PlayerID, payload.PlayerName))
					}

				case ActionUpdateResources, ActionDeductResources:
					if action.Meta.PlayerID != "" {
						eventBus.Publish(ctx, events.NewPlayerResourcesChangedEvent(action.Meta.GameID, action.Meta.PlayerID))
					}

				case ActionUpdateProduction:
					if action.Meta.PlayerID != "" {
						eventBus.Publish(ctx, events.NewPlayerProductionChangedEvent(action.Meta.GameID, action.Meta.PlayerID))
					}

				case ActionUpdateTerraformRating:
					if action.Meta.PlayerID != "" {
						eventBus.Publish(ctx, events.NewPlayerTRChangedEvent(action.Meta.GameID, action.Meta.PlayerID))
					}

				case ActionUpdateGlobalParams, ActionIncreaseTemperature, ActionIncreaseOxygen, ActionPlaceOcean:
					changeTypes := []string{}
					switch action.Type {
					case ActionIncreaseTemperature:
						changeTypes = append(changeTypes, "temperature")
					case ActionIncreaseOxygen:
						changeTypes = append(changeTypes, "oxygen")
					case ActionPlaceOcean:
						changeTypes = append(changeTypes, "oceans")
					default:
						changeTypes = append(changeTypes, "all")
					}
					eventBus.Publish(ctx, events.NewGlobalParametersChangedEvent(action.Meta.GameID, changeTypes))

				case ActionPlayCard:
					if payload, ok := action.Payload.(dto.PlayCardAction); ok {
						eventBus.Publish(ctx, events.NewCardPlayedEvent(action.Meta.GameID, action.Meta.PlayerID, payload.CardID))
					}

				case ActionSelectStartingCards:
					if payload, ok := action.Payload.(SelectStartingCardsPayload); ok {
						eventBus.Publish(ctx, events.NewCardSelectedEvent(payload.GameID, payload.PlayerID, payload.SelectedCards, payload.Cost))
					}
				}

				// Always emit a general game updated event for WebSocket broadcasting
				eventBus.Publish(ctx, events.NewGameUpdatedEvent(action.Meta.GameID))
			}

			return newState, nil
		}
	}
}

// ValidationMiddleware validates actions before they are applied
func ValidationMiddleware(next Reducer) Reducer {
	return func(state ApplicationState, action Action) (ApplicationState, error) {
		// Validate action based on current state
		if err := validateAction(state, action); err != nil {
			return state, err
		}

		return next(state, action)
	}
}

// validateAction performs business rule validation
func validateAction(state ApplicationState, action Action) error {
	// Basic validation - can be extended with more complex business rules
	switch action.Type {
	case ActionJoinGame:
		gameState, exists := state.GetGame(action.Meta.GameID)
		if !exists {
			return ErrGameNotFound
		}
		if gameState.Game().Status != "lobby" {
			return ErrGameNotJoinable
		}
		// Check if game is full
		// Add more validation as needed

	case ActionStartGame:
		gameState, exists := state.GetGame(action.Meta.GameID)
		if !exists {
			return ErrGameNotFound
		}
		if gameState.Game().HostPlayerID != action.Meta.PlayerID {
			return ErrNotHost
		}

	case ActionDeductResources:
		_, exists := state.GetPlayer(action.Meta.PlayerID)
		if !exists {
			return ErrPlayerNotFound
		}
		// Validate player has sufficient resources
		// Add resource validation logic

		// Add more validation cases as needed
	}

	return nil
}

// Error definitions
var (
	ErrGameNotFound          = NewStoreError("game not found")
	ErrPlayerNotFound        = NewStoreError("player not found")
	ErrGameNotJoinable       = NewStoreError("game is not joinable")
	ErrNotHost               = NewStoreError("only host can start game")
	ErrInsufficientResources = NewStoreError("insufficient resources")
)

// StoreError represents an error from the store
type StoreError struct {
	message string
}

func (e *StoreError) Error() string {
	return e.message
}

func NewStoreError(message string) *StoreError {
	return &StoreError{message: message}
}
