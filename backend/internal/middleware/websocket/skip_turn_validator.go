package websocketmiddleware

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/transaction"
)

// CreateSkipTurnValidatorMiddleware creates a middleware function that validates skip actions
// This is similar to turn validator but allows actions even when AvailableActions <= 0
func CreateSkipTurnValidatorMiddleware(transactionManager *transaction.Manager) MiddlewareFunc {
	return func(ctx context.Context, gameID, playerID string, actionRequest interface{}, next core.ActionHandler) error {
		// Use transaction system for atomic turn validation
		err := transactionManager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ValidateSkipTurn(gameID, playerID)
		})

		if err != nil {
			return fmt.Errorf("turn validation failed: %w", err)
		}

		// Validation passed, continue to next middleware/handler
		return next.Handle(ctx, gameID, playerID, actionRequest)
	}
}
