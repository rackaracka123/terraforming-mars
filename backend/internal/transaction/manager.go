package transaction

import (
	"context"

	"terraforming-mars-backend/internal/repository"
)

// Manager is the main entry point for creating and executing transactions
type Manager struct {
	playerRepo repository.PlayerRepository
	gameRepo   repository.GameRepository
}

// NewManager creates a new transaction manager
func NewManager(playerRepo repository.PlayerRepository, gameRepo repository.GameRepository) *Manager {
	return &Manager{
		playerRepo: playerRepo,
		gameRepo:   gameRepo,
	}
}

// GetPlayerRepo returns the player repository (implements TransactionManager interface)
func (m *Manager) GetPlayerRepo() repository.PlayerRepository {
	return m.playerRepo
}

// GetGameRepo returns the game repository (implements TransactionManager interface)
func (m *Manager) GetGameRepo() repository.GameRepository {
	return m.gameRepo
}

// ExecuteAtomic executes a function within an atomic transaction context
func (m *Manager) ExecuteAtomic(ctx context.Context, operations func(tx *Transaction) error) error {
	tx := NewTransaction(m)

	// Build operations within the transaction
	if err := operations(tx); err != nil {
		return err
	}

	// Execute all operations atomically
	if err := tx.Execute(ctx); err != nil {
		// Transaction automatically rolls back on error
		return err
	}

	return nil
}

// NewTransaction creates a new transaction (can be used for advanced cases)
func (m *Manager) NewTransaction() *Transaction {
	return NewTransaction(m)
}
