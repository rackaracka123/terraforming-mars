package core

import (
	"context"
)

// ActionHandler defines the interface that all action handlers must implement
type ActionHandler interface {
	Handle(ctx context.Context, gameID, playerID string, actionRequest interface{}) error
}

// ActionHandlerFunc is a function adapter that allows ordinary functions to be used as ActionHandlers
type ActionHandlerFunc func(ctx context.Context, gameID, playerID string, actionRequest interface{}) error

// Handle implements the ActionHandler interface for ActionHandlerFunc
func (f ActionHandlerFunc) Handle(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	return f(ctx, gameID, playerID, actionRequest)
}
