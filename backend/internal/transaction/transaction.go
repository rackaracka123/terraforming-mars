package transaction

import (
	"context"
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/transaction/operations"
)

// Transaction represents a single atomic transaction containing multiple operations
type Transaction struct {
	manager    TransactionManager
	operations []Operation
	rolledBack bool
	committed  bool
	mutex      sync.RWMutex
}

// NewTransaction creates a new transaction
func NewTransaction(manager TransactionManager) *Transaction {
	return &Transaction{
		manager:    manager,
		operations: make([]Operation, 0),
		rolledBack: false,
		committed:  false,
	}
}

// AddOperation adds an operation to the transaction
func (t *Transaction) AddOperation(op Operation) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.committed || t.rolledBack {
		return // Cannot add operations to finished transaction
	}

	t.operations = append(t.operations, op)
}

// Execute runs all operations in the transaction
func (t *Transaction) Execute(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.committed || t.rolledBack {
		return fmt.Errorf("transaction already finished")
	}

	// Execute all operations
	for i, op := range t.operations {
		if err := op.Execute(ctx); err != nil {
			// Rollback all previously executed operations
			for j := i - 1; j >= 0; j-- {
				if rollbackErr := t.operations[j].Rollback(ctx); rollbackErr != nil {
					// Log rollback error but continue rolling back
					// In a real system, this would use the logger
				}
			}
			t.rolledBack = true
			return fmt.Errorf("operation %d (%s) failed: %w", i, op.String(), err)
		}
	}

	t.committed = true
	return nil
}

// Rollback undoes all executed operations
func (t *Transaction) Rollback(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.rolledBack {
		return nil // Already rolled back
	}

	if t.committed {
		return fmt.Errorf("cannot rollback committed transaction")
	}

	// Rollback all operations in reverse order
	var rollbackErrors []error
	for i := len(t.operations) - 1; i >= 0; i-- {
		if err := t.operations[i].Rollback(ctx); err != nil {
			rollbackErrors = append(rollbackErrors, err)
		}
	}

	t.rolledBack = true

	if len(rollbackErrors) > 0 {
		return fmt.Errorf("rollback completed with errors: %v", rollbackErrors)
	}

	return nil
}

// IsCommitted returns whether the transaction has been committed
func (t *Transaction) IsCommitted() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.committed
}

// IsRolledBack returns whether the transaction has been rolled back
func (t *Transaction) IsRolledBack() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.rolledBack
}

// ValidateTurn creates an operation to validate a player's turn
func (t *Transaction) ValidateTurn(gameID, playerID string) error {
	op := operations.NewValidateTurnOperation(t.manager.GetGameRepo(), t.manager.GetPlayerRepo(), gameID, playerID)
	t.AddOperation(op)
	return nil
}

// ValidateResources creates an operation to validate resources without deducting them
func (t *Transaction) ValidateResources(gameID, playerID string, cost model.Resources) error {
	op := operations.NewValidateResourcesOperation(t.manager.GetPlayerRepo(), gameID, playerID, cost)
	t.AddOperation(op)
	return nil
}

// DeductResources creates an operation to deduct resources from a player (for purchases)
func (t *Transaction) DeductResources(gameID, playerID string, cost model.Resources) error {
	op := operations.NewDeductResourcesOperation(t.manager.GetPlayerRepo(), gameID, playerID, cost)
	t.AddOperation(op)
	return nil
}

// ConsumePlayerAction creates an operation to consume a player action (for turn actions)
func (t *Transaction) ConsumePlayerAction(gameID, playerID string) error {
	op := operations.NewConsumeActionOperation(t.manager.GetPlayerRepo(), gameID, playerID)
	t.AddOperation(op)
	return nil
}

// ProcessTurnAction combines turn validation, resource deduction, and action consumption
func (t *Transaction) ProcessTurnAction(gameID, playerID string, cost model.Resources) error {
	// Validate turn first
	if err := t.ValidateTurn(gameID, playerID); err != nil {
		return err
	}

	// Deduct resources if there's a cost
	if !cost.IsZero() {
		if err := t.DeductResources(gameID, playerID, cost); err != nil {
			return err
		}
	}

	// Consume one action
	if err := t.ConsumePlayerAction(gameID, playerID); err != nil {
		return err
	}

	return nil
}

// ProcessPurchase combines resource validation and deduction (no action consumed)
func (t *Transaction) ProcessPurchase(gameID, playerID string, cost model.Resources) error {
	// Just deduct resources for purchases
	return t.DeductResources(gameID, playerID, cost)
}
