package websocketmiddleware

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/transaction"
)

// CreateActionValidatorMiddleware creates a middleware function that validates the player has available actions
// and transactionally consumes one action if the request succeeds
func CreateActionValidatorMiddleware(transactionManager *transaction.Manager) MiddlewareFunc {
	return func(ctx context.Context, gameID, playerID string, actionRequest interface{}, next core.ActionHandler) error {
		// Execute the handler within a transaction that includes action consumption
		return transactionManager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			// First consume the action (this validates the player has actions available)
			if err := tx.ConsumePlayerAction(gameID, playerID); err != nil {
				return fmt.Errorf("action validation failed: %w", err)
			}

			// Execute the actual handler - if this fails, the action consumption will be rolled back
			return next.Handle(ctx, gameID, playerID, actionRequest)
		})
	}
}
