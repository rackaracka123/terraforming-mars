package core

import (
	"context"
)

// ActionHandler defines the interface that all action handlers must implement
type ActionHandler interface {
	Handle(ctx context.Context, gameID, playerID string, actionRequest interface{}) error
}
