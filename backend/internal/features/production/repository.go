package production

import (
	"context"
	"sync"
)

// Repository manages production phase state storage
type Repository interface {
	// Get state
	Get(ctx context.Context) (*ProductionPhaseState, error)
	IsSelectionComplete(ctx context.Context) (bool, error)
	GetAvailableCards(ctx context.Context) ([]string, error)

	// Granular operations
	SetAvailableCards(ctx context.Context, cardIDs []string) error
	MarkSelectionComplete(ctx context.Context) error
	ClearState(ctx context.Context) error
}

// RepositoryImpl implements independent in-memory storage for production phase state
type RepositoryImpl struct {
	mu    sync.RWMutex
	state *ProductionPhaseState
}

// NewRepository creates a new independent production phase state repository with initial state
func NewRepository(initialState *ProductionPhaseState) Repository {
	var stateCopy *ProductionPhaseState
	if initialState != nil {
		copy := *initialState
		copy.AvailableCards = append([]string{}, initialState.AvailableCards...)
		stateCopy = &copy
	}
	return &RepositoryImpl{
		state: stateCopy,
	}
}

// Get retrieves current production phase state
func (r *RepositoryImpl) Get(ctx context.Context) (*ProductionPhaseState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.state == nil {
		return nil, nil
	}

	// Return a copy
	stateCopy := *r.state
	stateCopy.AvailableCards = append([]string{}, r.state.AvailableCards...)
	return &stateCopy, nil
}

// IsSelectionComplete checks if selection is complete
func (r *RepositoryImpl) IsSelectionComplete(ctx context.Context) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.state == nil {
		return false, nil
	}

	return r.state.SelectionComplete, nil
}

// GetAvailableCards retrieves available cards for selection
func (r *RepositoryImpl) GetAvailableCards(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.state == nil {
		return []string{}, nil
	}

	return append([]string{}, r.state.AvailableCards...), nil
}

// SetAvailableCards sets the cards available for selection
func (r *RepositoryImpl) SetAvailableCards(ctx context.Context, cardIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state == nil {
		r.state = &ProductionPhaseState{
			AvailableCards:    append([]string{}, cardIDs...),
			SelectionComplete: false,
		}
	} else {
		r.state.AvailableCards = append([]string{}, cardIDs...)
	}

	return nil
}

// MarkSelectionComplete marks the selection as complete
func (r *RepositoryImpl) MarkSelectionComplete(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state == nil {
		r.state = &ProductionPhaseState{
			AvailableCards:    []string{},
			SelectionComplete: true,
		}
	} else {
		r.state.SelectionComplete = true
	}

	return nil
}

// ClearState clears the production phase state
func (r *RepositoryImpl) ClearState(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.state = nil
	return nil
}
