package repository

import (
	"fmt"
	"sync"
	"terraforming-mars-backend/internal/domain"
	"time"
)

// GameRepository handles game state persistence and retrieval
type GameRepository struct {
	games  map[string]*domain.GameState
	mutex  sync.RWMutex
}

// NewGameRepository creates a new game repository
func NewGameRepository() *GameRepository {
	repo := &GameRepository{
		games: make(map[string]*domain.GameState),
		mutex: sync.RWMutex{},
	}
	
	// Create demo game
	demoGame := createDemoGame()
	repo.games[demoGame.ID] = demoGame
	
	return repo
}

// GetGame retrieves a game by ID
func (r *GameRepository) GetGame(id string) (*domain.GameState, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	game, exists := r.games[id]
	if !exists {
		return nil, fmt.Errorf("game with ID %s not found", id)
	}
	
	return game, nil
}

// SaveGame saves or updates a game
func (r *GameRepository) SaveGame(game *domain.GameState) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	game.UpdatedAt = time.Now()
	r.games[game.ID] = game
	
	return nil
}

// DeleteGame removes a game
func (r *GameRepository) DeleteGame(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	delete(r.games, id)
	return nil
}

// ListGames returns all games
func (r *GameRepository) ListGames() ([]*domain.GameState, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	games := make([]*domain.GameState, 0, len(r.games))
	for _, game := range r.games {
		games = append(games, game)
	}
	
	return games, nil
}

// createDemoGame creates the initial demo game
func createDemoGame() *domain.GameState {
	now := time.Now()
	
	return &domain.GameState{
		ID:            "demo",
		Players:       []domain.Player{},
		CurrentPlayer: "",
		Generation:    1,
		Phase:         domain.GamePhaseInitialResearch,
		GlobalParameters: domain.GlobalParameters{
			Temperature: 2,
			Oxygen:      7,
			Oceans:      0,
		},
		Milestones:         []domain.Milestone{},
		Awards:             []domain.Award{},
		FirstPlayer:        "",
		Deck:               []string{},
		DiscardPile:        []string{},
		SoloMode:           false,
		Turn:               1,
		GameSettings:       getDefaultGameSettings(),
		CurrentActionCount: intPtr(0),
		MaxActionsPerTurn:  intPtr(2),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// getDefaultGameSettings returns default game settings
func getDefaultGameSettings() domain.GameSettings {
	return domain.GameSettings{
		Expansions:                    []domain.GameExpansion{},
		CorporateEra:                  false,
		DraftVariant:                  false,
		InitialDraft:                  false,
		PreludeExtension:              false,
		VenusNextExtension:            false,
		ColoniesExtension:             false,
		TurmoilExtension:              false,
		RemoveNegativeAttackCards:     false,
		IncludeVenusMA:                false,
		MoonExpansion:                 false,
		PathfindersExpansion:          false,
		UnderworldExpansion:           false,
		EscapeVelocityExpansion:       false,
		Fast:                          false,
		ShowOtherPlayersVP:            true,
		SoloTR:                        false,
		RandomFirstPlayer:             false,
		RequiresVenusTrackCompletion:  false,
		RequiresMoonTrackCompletion:   false,
		MoonStandardProjectVariant:    false,
		AltVenusBoard:                 false,
		EscapeVelocityMode:            false,
		EscapeVelocityThreshold:       30,
		EscapeVelocityPeriod:          2,
		EscapeVelocityPenalty:         1,
		TwoTempTerraformingThreshold:  false,
		HeatFor:                       false,
		Breakthrough:                  false,
	}
}

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}