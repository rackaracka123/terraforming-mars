package repository

import (
	"context"
	"sync"
)

// CardSelectionRepository manages player card selections during selection phases
// This repository is ephemeral - it resets after each selection phase
type CardSelectionRepository interface {
	// StorePlayerOptions stores the card options available to a player
	StorePlayerOptions(ctx context.Context, gameID, playerID string, cardOptions []string) error

	// GetPlayerOptions returns the card options for a player
	GetPlayerOptions(ctx context.Context, gameID, playerID string) ([]string, error)

	// MarkSelectionComplete marks a player as having completed their selection
	MarkSelectionComplete(ctx context.Context, gameID, playerID string) error

	// IsSelectionComplete checks if a player has completed their selection
	IsSelectionComplete(ctx context.Context, gameID, playerID string) (bool, error)

	// IsAllSelectionsComplete checks if all players have completed their selections
	IsAllSelectionsComplete(ctx context.Context, gameID string) (bool, error)

	// Clear clears all selection data for a game (called after each selection phase)
	Clear(ctx context.Context, gameID string) error
}

// CardSelectionRepositoryImpl implements CardSelectionRepository
type CardSelectionRepositoryImpl struct {
	mutex sync.RWMutex
	// Store card options and completion status by game
	playerCardOptions map[string]map[string][]string // gameID -> playerID -> cardOptions
	selectionStatus   map[string]map[string]bool     // gameID -> playerID -> hasSelected
}

// NewCardSelectionRepository creates a new card selection repository
func NewCardSelectionRepository() CardSelectionRepository {
	return &CardSelectionRepositoryImpl{
		playerCardOptions: make(map[string]map[string][]string),
		selectionStatus:   make(map[string]map[string]bool),
	}
}

// StorePlayerOptions stores the card options available to a player
func (r *CardSelectionRepositoryImpl) StorePlayerOptions(ctx context.Context, gameID, playerID string, cardOptions []string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.playerCardOptions[gameID] == nil {
		r.playerCardOptions[gameID] = make(map[string][]string)
	}
	if r.selectionStatus[gameID] == nil {
		r.selectionStatus[gameID] = make(map[string]bool)
	}

	r.playerCardOptions[gameID][playerID] = cardOptions
	r.selectionStatus[gameID][playerID] = false

	return nil
}

// GetPlayerOptions returns the card options for a player
func (r *CardSelectionRepositoryImpl) GetPlayerOptions(ctx context.Context, gameID, playerID string) ([]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if gameOptions, exists := r.playerCardOptions[gameID]; exists {
		if playerOptions, exists := gameOptions[playerID]; exists {
			return playerOptions, nil
		}
	}
	return nil, nil
}

// MarkSelectionComplete marks a player as having completed their selection
func (r *CardSelectionRepositoryImpl) MarkSelectionComplete(ctx context.Context, gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.selectionStatus[gameID] == nil {
		r.selectionStatus[gameID] = make(map[string]bool)
	}
	r.selectionStatus[gameID][playerID] = true

	return nil
}

// IsSelectionComplete checks if a player has completed their selection
func (r *CardSelectionRepositoryImpl) IsSelectionComplete(ctx context.Context, gameID, playerID string) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if gameStatus, exists := r.selectionStatus[gameID]; exists {
		if completed, exists := gameStatus[playerID]; exists {
			return completed, nil
		}
	}
	return false, nil
}

// IsAllSelectionsComplete checks if all players have completed their selections
func (r *CardSelectionRepositoryImpl) IsAllSelectionsComplete(ctx context.Context, gameID string) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	gameStatus, exists := r.selectionStatus[gameID]
	if !exists {
		return false, nil
	}

	// Check if all players who have options have completed selection
	gameOptions, exists := r.playerCardOptions[gameID]
	if !exists {
		return false, nil
	}

	for playerID := range gameOptions {
		if completed, exists := gameStatus[playerID]; !exists || !completed {
			return false, nil
		}
	}

	return true, nil
}

// Clear clears all selection data for a game
// Called after each selection phase (starting cards, production phase cards, etc.)
func (r *CardSelectionRepositoryImpl) Clear(ctx context.Context, gameID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.playerCardOptions, gameID)
	delete(r.selectionStatus, gameID)

	return nil
}
