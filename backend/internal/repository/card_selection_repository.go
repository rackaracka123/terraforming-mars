package repository

import (
	"fmt"
	"sync"
	"terraforming-mars-backend/internal/model/card_selection"
)

// CardSelectionRepository manages card selection data for different phases
type CardSelectionRepository struct {
	startingCardSelections map[string]*card_selection.CardSelection
	mutex                  sync.RWMutex
}

// NewCardSelectionRepository creates a new card selection repository
func NewCardSelectionRepository() *CardSelectionRepository {
	return &CardSelectionRepository{
		startingCardSelections: make(map[string]*card_selection.CardSelection),
	}
}

// CreateStartingCardSelection creates a new starting card selection for a game
func (r *CardSelectionRepository) CreateStartingCardSelection(gameID string, playerCardOptions []card_selection.PlayerCardOptions) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	selection := &card_selection.CardSelection{
		PlayerCardOptions:            playerCardOptions,
		PlayersWhoCompletedSelection: make([]string, 0),
	}

	r.startingCardSelections[gameID] = selection
	return nil
}

// GetStartingCardSelection gets the starting card selection for a game
func (r *CardSelectionRepository) GetStartingCardSelection(gameID string) (*card_selection.CardSelection, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	selection, exists := r.startingCardSelections[gameID]
	if !exists {
		return nil, fmt.Errorf("starting card selection not found for game: %s", gameID)
	}

	return selection, nil
}

// GetPlayerStartingCardOptions gets the starting card options for a specific player
func (r *CardSelectionRepository) GetPlayerStartingCardOptions(gameID, playerID string) ([]string, error) {
	selection, err := r.GetStartingCardSelection(gameID)
	if err != nil {
		return nil, err
	}

	for _, playerOptions := range selection.PlayerCardOptions {
		if playerOptions.PlayerID == playerID {
			return playerOptions.CardOptions, nil
		}
	}

	return nil, fmt.Errorf("starting card options not found for player %s in game %s", playerID, gameID)
}

// MarkPlayerCompletedStartingCardSelection marks a player as having completed their starting card selection
func (r *CardSelectionRepository) MarkPlayerCompletedStartingCardSelection(gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	selection, exists := r.startingCardSelections[gameID]
	if !exists {
		return fmt.Errorf("starting card selection not found for game: %s", gameID)
	}

	// Check if already marked
	for _, selectedPlayerID := range selection.PlayersWhoCompletedSelection {
		if selectedPlayerID == playerID {
			return nil // Already marked
		}
	}

	selection.PlayersWhoCompletedSelection = append(selection.PlayersWhoCompletedSelection, playerID)
	return nil
}

// AllPlayersCompletedStartingCardSelection checks if all players have completed their starting card selection
func (r *CardSelectionRepository) AllPlayersCompletedStartingCardSelection(gameID string, totalPlayers int) (bool, error) {
	selection, err := r.GetStartingCardSelection(gameID)
	if err != nil {
		return false, err
	}

	return len(selection.PlayersWhoCompletedSelection) >= totalPlayers, nil
}

// DeleteStartingCardSelection removes the starting card selection data for a game (cleanup after phase ends)
func (r *CardSelectionRepository) DeleteStartingCardSelection(gameID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.startingCardSelections, gameID)
	return nil
}