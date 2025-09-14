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
		// Create a transaction to handle action consumption atomically
		tx := transactionManager.NewTransaction()

		// Add the action consumption operation to the transaction
		if err := tx.ConsumePlayerAction(ctx, gameID, playerID); err != nil {
			return fmt.Errorf("action validation failed: %w", err)
		}

		// Execute the actual handler first - if this fails, we won't execute the transaction
		if err := next.Handle(ctx, gameID, playerID, actionRequest); err != nil {
			// Handler failed - don't execute the transaction (no action consumption)
			return err
		}

		// Handler succeeded - execute the transaction to consume the action
		if err := tx.Execute(ctx); err != nil {
			// This should rarely happen since we already validated the action, but handle it
			return fmt.Errorf("failed to consume action: %w", err)
		}

		return nil
	}
}
