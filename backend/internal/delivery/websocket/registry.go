package websocket

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_aquifer"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_city"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_power_plant"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/launch_asteroid"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/plant_greenery"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/play_card"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/sell_patents"
	"terraforming-mars-backend/internal/delivery/websocket/handler/card_selection/select_cards"
	"terraforming-mars-backend/internal/delivery/websocket/handler/card_selection/select_starting_card"
	"terraforming-mars-backend/internal/delivery/websocket/handler/connect"
	"terraforming-mars-backend/internal/delivery/websocket/handler/game/skip_action"
	"terraforming-mars-backend/internal/delivery/websocket/handler/game/start_game"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	websocketmiddleware "terraforming-mars-backend/internal/middleware/websocket"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/internal/transaction"
)

// RegisterHandlers registers all message type handlers with the hub
func RegisterHandlers(hub *core.Hub, gameService service.GameService, playerService service.PlayerService, standardProjectService service.StandardProjectService, cardService service.CardService, gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) {
	parser := utils.NewMessageParser()

	// Create transaction manager for middleware
	transactionManager := transaction.NewManager(playerRepo, gameRepo)

	// Create turn validation middleware
	turnValidationMiddleware := websocketmiddleware.CreateTurnValidatorMiddleware(transactionManager)

	// Create skip turn validation middleware (allows skip with 0 actions)
	skipTurnValidationMiddleware := websocketmiddleware.CreateSkipTurnValidatorMiddleware(transactionManager)

	// Register connection handler (no middleware needed - not turn-based)
	connectionHandler := connect.NewConnectionHandler(gameService, playerService)
	hub.RegisterHandler(dto.MessageTypePlayerConnect, connectionHandler)

	// Register standard project handlers WITH middleware for turn-based actions
	hub.RegisterHandler(dto.MessageTypeActionLaunchAsteroid, wrapWithTurnValidation(launch_asteroid.NewHandler(standardProjectService), turnValidationMiddleware))
	hub.RegisterHandler(dto.MessageTypeActionSellPatents, wrapWithTurnValidation(sell_patents.NewHandler(standardProjectService, parser), turnValidationMiddleware))
	hub.RegisterHandler(dto.MessageTypeActionBuildPowerPlant, wrapWithTurnValidation(build_power_plant.NewHandler(standardProjectService), turnValidationMiddleware))
	hub.RegisterHandler(dto.MessageTypeActionBuildAquifer, wrapWithTurnValidation(build_aquifer.NewHandler(standardProjectService, parser), turnValidationMiddleware))
	hub.RegisterHandler(dto.MessageTypeActionPlantGreenery, wrapWithTurnValidation(plant_greenery.NewHandler(standardProjectService, parser), turnValidationMiddleware))
	hub.RegisterHandler(dto.MessageTypeActionBuildCity, wrapWithTurnValidation(build_city.NewHandler(standardProjectService, parser), turnValidationMiddleware))

	// Skip action needs special validation that allows 0 actions
	hub.RegisterHandler(dto.MessageTypeActionSkipAction, wrapWithTurnValidation(skip_action.NewHandler(gameService, playerService), skipTurnValidationMiddleware))

	// Register game management handlers WITHOUT middleware (not turn-based)
	hub.RegisterHandler(dto.MessageTypeActionStartGame, start_game.NewHandler(gameService))

	// Register card selection handlers (may need turn validation depending on game phase)
	hub.RegisterHandler(dto.MessageTypeActionSelectStartingCard, select_starting_card.NewHandler(cardService, gameService, hub.GetSessionManager(), parser))
	hub.RegisterHandler(dto.MessageTypeActionSelectCards, select_cards.NewHandler(cardService, gameService, parser))

	// Register play card handler WITH turn validation ONLY (action consumption handled in card service after validation)
	hub.RegisterHandler(dto.MessageTypeActionPlayCard, wrapWithTurnValidation(play_card.NewHandler(cardService, parser), turnValidationMiddleware))
}

// wrapWithTurnValidation wraps a MessageHandler with turn validation middleware
func wrapWithTurnValidation(handler core.MessageHandler, turnValidation websocketmiddleware.MiddlewareFunc) core.MessageHandler {
	return &middlewareWrapper{
		handler:        handler,
		turnValidation: turnValidation,
	}
}

// wrapWithTurnAndActionValidation wraps a MessageHandler with both turn and action validation middleware
func wrapWithTurnAndActionValidation(handler core.MessageHandler, turnValidation websocketmiddleware.MiddlewareFunc, actionValidation websocketmiddleware.MiddlewareFunc) core.MessageHandler {
	return &actionMiddlewareWrapper{
		handler:          handler,
		turnValidation:   turnValidation,
		actionValidation: actionValidation,
	}
}

// middlewareWrapper adapts MessageHandler to work with ActionHandler middleware
type middlewareWrapper struct {
	handler        core.MessageHandler
	turnValidation websocketmiddleware.MiddlewareFunc
}

// HandleMessage implements MessageHandler interface with middleware
func (w *middlewareWrapper) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		// Let the wrapped handler handle the error
		w.handler.HandleMessage(ctx, connection, message)
		return
	}

	// Create an ActionHandler adapter for the original MessageHandler
	actionHandler := core.ActionHandlerFunc(func(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
		// Call the original MessageHandler
		w.handler.HandleMessage(ctx, connection, message)
		return nil
	})

	// Apply turn validation middleware
	err := w.turnValidation(ctx, gameID, playerID, message.Payload, actionHandler)
	if err != nil {
		// Send error back to client
		errorHandler := utils.NewErrorHandler()
		errorHandler.SendError(connection, "Turn validation failed: "+err.Error())
		return
	}
}

// actionMiddlewareWrapper adapts MessageHandler to work with both turn and action validation middleware
type actionMiddlewareWrapper struct {
	handler          core.MessageHandler
	turnValidation   websocketmiddleware.MiddlewareFunc
	actionValidation websocketmiddleware.MiddlewareFunc
}

// HandleMessage implements MessageHandler interface with both turn and action validation middleware
func (w *actionMiddlewareWrapper) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		// Let the wrapped handler handle the error
		w.handler.HandleMessage(ctx, connection, message)
		return
	}

	// Create an ActionHandler adapter for the original MessageHandler
	actionHandler := core.ActionHandlerFunc(func(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
		// Call the original MessageHandler
		w.handler.HandleMessage(ctx, connection, message)
		return nil
	})

	// Chain middlewares: turn validation first, then action validation
	chainedHandler := core.ActionHandlerFunc(func(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
		// Apply action validation middleware
		return w.actionValidation(ctx, gameID, playerID, actionRequest, actionHandler)
	})

	// Apply turn validation middleware first
	err := w.turnValidation(ctx, gameID, playerID, message.Payload, chainedHandler)
	if err != nil {
		// Send error back to client
		errorHandler := utils.NewErrorHandler()
		errorHandler.SendError(connection, "Validation failed: "+err.Error())
		return
	}
}
