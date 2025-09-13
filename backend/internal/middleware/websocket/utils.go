package websocketmiddleware

import (
	"context"

	"terraforming-mars-backend/internal/delivery/websocket/core"
)

// MiddlewareFunc is a function adapter for middleware functionality
type MiddlewareFunc func(ctx context.Context, gameID, playerID string, actionRequest interface{}, next core.ActionHandler) error

// ChainMiddleware chains multiple middleware functions together
func ChainMiddleware(middlewares ...MiddlewareFunc) MiddlewareFunc {
	return func(ctx context.Context, gameID, playerID string, actionRequest interface{}, next core.ActionHandler) error {
		// Build the chain from right to left
		handler := next
		for i := len(middlewares) - 1; i >= 0; i-- {
			middleware := middlewares[i]
			currentHandler := handler
			handler = core.ActionHandlerFunc(func(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
				return middleware(ctx, gameID, playerID, actionRequest, currentHandler)
			})
		}
		return handler.Handle(ctx, gameID, playerID, actionRequest)
	}
}

// WrapWithMiddleware wraps an action handler with middleware
func WrapWithMiddleware(handler core.ActionHandler, middleware MiddlewareFunc) core.ActionHandler {
	return core.ActionHandlerFunc(func(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
		return middleware(ctx, gameID, playerID, actionRequest, handler)
	})
}
