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
	broadcaster := hub.GetBroadcaster()
	manager := hub.GetManager()

	// Create transaction manager for middleware
	transactionManager := transaction.NewManager(playerRepo, gameRepo)

	// Create turn validation middleware
	turnValidationMiddleware := websocketmiddleware.CreateTurnValidatorMiddleware(transactionManager)

	// Register connection handler (no middleware needed - not turn-based)
	connectionHandler := connect.NewConnectionHandler(gameService, playerService, broadcaster, manager)
	hub.RegisterHandler(dto.MessageTypePlayerConnect, connectionHandler)

	// Register standard project handlers WITH middleware for turn-based actions
	hub.RegisterHandler(dto.MessageTypeActionLaunchAsteroid, wrapWithTurnValidation(launch_asteroid.NewHandler(standardProjectService), turnValidationMiddleware))
	hub.RegisterHandler(dto.MessageTypeActionSellPatents, wrapWithTurnValidation(sell_patents.NewHandler(standardProjectService, parser), turnValidationMiddleware))
	hub.RegisterHandler(dto.MessageTypeActionBuildPowerPlant, wrapWithTurnValidation(build_power_plant.NewHandler(standardProjectService), turnValidationMiddleware))
	hub.RegisterHandler(dto.MessageTypeActionBuildAquifer, wrapWithTurnValidation(build_aquifer.NewHandler(standardProjectService, parser), turnValidationMiddleware))
	hub.RegisterHandler(dto.MessageTypeActionPlantGreenery, wrapWithTurnValidation(plant_greenery.NewHandler(standardProjectService, parser), turnValidationMiddleware))
	hub.RegisterHandler(dto.MessageTypeActionBuildCity, wrapWithTurnValidation(build_city.NewHandler(standardProjectService, parser), turnValidationMiddleware))

	// Skip action needs turn validation too (player must be on their turn to skip)
	hub.RegisterHandler(dto.MessageTypeActionSkipAction, wrapWithTurnValidation(skip_action.NewHandler(gameService, playerService, broadcaster), turnValidationMiddleware))

	// Register game management handlers WITHOUT middleware (not turn-based)
	hub.RegisterHandler(dto.MessageTypeActionStartGame, start_game.NewHandler(gameService))

	// Register card selection handlers (may need turn validation depending on game phase)
	hub.RegisterHandler(dto.MessageTypeActionSelectStartingCard, select_starting_card.NewHandler(cardService, gameService, parser))
	hub.RegisterHandler(dto.MessageTypeActionSelectCards, select_cards.NewHandler(cardService, gameService, parser))

	// TODO: Register play card handler when it's created
	// hub.RegisterHandler(dto.MessageTypeActionPlayCard, play_card.NewHandler(...))
}

// wrapWithTurnValidation wraps a MessageHandler with turn validation middleware
func wrapWithTurnValidation(handler core.MessageHandler, turnValidation websocketmiddleware.MiddlewareFunc) core.MessageHandler {
	return &middlewareWrapper{
		handler:        handler,
		turnValidation: turnValidation,
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
