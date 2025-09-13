package transaction

import (
	"context"

	"terraforming-mars-backend/internal/repository"
)

// Operation represents a single atomic operation that can be rolled back
type Operation interface {
	Execute(ctx context.Context) error
	Rollback(ctx context.Context) error
	String() string
}

// TransactionManager interface defines the manager that transactions depend on
type TransactionManager interface {
	GetPlayerRepo() repository.PlayerRepository
	GetGameRepo() repository.GameRepository
}
