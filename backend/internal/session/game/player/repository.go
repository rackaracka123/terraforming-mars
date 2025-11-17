package player

import (
	"context"
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
)

// Repository manages player data with event-driven updates
type Repository interface {
	// Create creates a new player in a game
	Create(ctx context.Context, gameID string, player *Player) error

	// GetByID retrieves a player by ID from a specific game
	GetByID(ctx context.Context, gameID string, playerID string) (*Player, error)

	// ListByGameID retrieves all players in a game
	ListByGameID(ctx context.Context, gameID string) ([]*Player, error)

	// UpdateResources updates player resources (event-driven)
	UpdateResources(ctx context.Context, gameID string, playerID string, resources model.Resources) error

	// UpdateConnectionStatus updates player connection status
	UpdateConnectionStatus(ctx context.Context, gameID string, playerID string, isConnected bool) error

	// SetStartingCardsSelection sets the starting cards selection phase for a player
	SetStartingCardsSelection(ctx context.Context, gameID string, playerID string, cardIDs []string, corpIDs []string) error

	// AddCard adds a card to player's hand
	AddCard(ctx context.Context, gameID string, playerID string, cardID string) error

	// SetCorporation sets the player's corporation
	SetCorporation(ctx context.Context, gameID string, playerID string, corporationID string) error

	// CompleteStartingSelection marks the starting selection as complete
	CompleteStartingSelection(ctx context.Context, gameID string, playerID string) error

	// CompleteProductionSelection marks the production selection as complete
	CompleteProductionSelection(ctx context.Context, gameID string, playerID string) error

	// UpdateProduction updates player production
	UpdateProduction(ctx context.Context, gameID string, playerID string, production model.Production) error

	// UpdateSelectStartingCardsPhase updates the starting cards selection phase
	UpdateSelectStartingCardsPhase(ctx context.Context, gameID string, playerID string, phase *SelectStartingCardsPhase) error

	// UpdateProductionPhase updates the production phase
	UpdateProductionPhase(ctx context.Context, gameID string, playerID string, phase *model.ProductionPhase) error

	// UpdateCorporation updates the player's corporation with full card data
	UpdateCorporation(ctx context.Context, gameID string, playerID string, corporation model.Card) error

	// UpdatePaymentSubstitutes updates player payment substitutes
	UpdatePaymentSubstitutes(ctx context.Context, gameID string, playerID string, substitutes []model.PaymentSubstitute) error

	// UpdatePlayerActions updates player actions
	UpdatePlayerActions(ctx context.Context, gameID string, playerID string, actions []model.PlayerAction) error

	// UpdateForcedFirstAction updates player forced first action
	UpdateForcedFirstAction(ctx context.Context, gameID string, playerID string, action *model.ForcedFirstAction) error
}

// RepositoryImpl implements the Repository interface with in-memory storage
type RepositoryImpl struct {
	mu       sync.RWMutex
	players  map[string]map[string]*Player // gameID -> playerID -> Player
	eventBus *events.EventBusImpl
}

// NewRepository creates a new player repository
func NewRepository(eventBus *events.EventBusImpl) Repository {
	return &RepositoryImpl{
		players:  make(map[string]map[string]*Player),
		eventBus: eventBus,
	}
}

// Create creates a new player in a game
func (r *RepositoryImpl) Create(ctx context.Context, gameID string, player *Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.players[gameID]; !exists {
		r.players[gameID] = make(map[string]*Player)
	}

	if _, exists := r.players[gameID][player.ID]; exists {
		return fmt.Errorf("player %s already exists in game %s", player.ID, gameID)
	}

	r.players[gameID][player.ID] = player

	// Event publishing can be added here if needed

	return nil
}

// GetByID retrieves a player by ID from a specific game
func (r *RepositoryImpl) GetByID(ctx context.Context, gameID string, playerID string) (*Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return nil, &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return nil, &model.NotFoundError{Resource: "player", ID: playerID}
	}

	return player, nil
}

// ListByGameID retrieves all players in a game
func (r *RepositoryImpl) ListByGameID(ctx context.Context, gameID string) ([]*Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return []*Player{}, nil
	}

	result := make([]*Player, 0, len(gamePlayers))
	for _, player := range gamePlayers {
		result = append(result, player)
	}

	return result, nil
}

// UpdateResources updates player resources (event-driven)
func (r *RepositoryImpl) UpdateResources(ctx context.Context, gameID string, playerID string, resources model.Resources) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.Resources = resources

	// Event publishing can be added here if needed
	// For now, simplified for proof of concept

	return nil
}

// UpdateConnectionStatus updates player connection status
func (r *RepositoryImpl) UpdateConnectionStatus(ctx context.Context, gameID string, playerID string, isConnected bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.IsConnected = isConnected

	return nil
}

// SetStartingCardsSelection sets the starting cards selection phase for a player
func (r *RepositoryImpl) SetStartingCardsSelection(ctx context.Context, gameID string, playerID string, cardIDs []string, corpIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.SelectStartingCardsPhase = &SelectStartingCardsPhase{
		AvailableCards:        cardIDs,
		AvailableCorporations: corpIDs,
		SelectionComplete:     false,
	}

	return nil
}

// AddCard adds a card to player's hand
func (r *RepositoryImpl) AddCard(ctx context.Context, gameID string, playerID string, cardID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.Cards = append(player.Cards, cardID)

	return nil
}

// SetCorporation sets the player's corporation
func (r *RepositoryImpl) SetCorporation(ctx context.Context, gameID string, playerID string, corporationID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.CorporationID = corporationID

	return nil
}

// CompleteStartingSelection marks the starting selection as complete
func (r *RepositoryImpl) CompleteStartingSelection(ctx context.Context, gameID string, playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	if player.SelectStartingCardsPhase != nil {
		player.SelectStartingCardsPhase.SelectionComplete = true
	}

	return nil
}

// CompleteProductionSelection marks the production selection as complete
func (r *RepositoryImpl) CompleteProductionSelection(ctx context.Context, gameID string, playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	if player.ProductionPhase != nil {
		player.ProductionPhase.SelectionComplete = true
	}

	return nil
}

// UpdateProduction updates player production
func (r *RepositoryImpl) UpdateProduction(ctx context.Context, gameID string, playerID string, production model.Production) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.Production = production

	return nil
}

// UpdateSelectStartingCardsPhase updates the starting cards selection phase
func (r *RepositoryImpl) UpdateSelectStartingCardsPhase(ctx context.Context, gameID string, playerID string, phase *SelectStartingCardsPhase) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.SelectStartingCardsPhase = phase

	return nil
}

// UpdateProductionPhase updates the production phase
func (r *RepositoryImpl) UpdateProductionPhase(ctx context.Context, gameID string, playerID string, phase *model.ProductionPhase) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.ProductionPhase = phase

	return nil
}

// UpdateCorporation updates the player's corporation with full card data
func (r *RepositoryImpl) UpdateCorporation(ctx context.Context, gameID string, playerID string, corporation model.Card) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.Corporation = &corporation
	player.CorporationID = corporation.ID

	return nil
}

// UpdatePaymentSubstitutes updates player payment substitutes
func (r *RepositoryImpl) UpdatePaymentSubstitutes(ctx context.Context, gameID string, playerID string, substitutes []model.PaymentSubstitute) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.PaymentSubstitutes = substitutes

	return nil
}

// UpdatePlayerActions updates player actions
func (r *RepositoryImpl) UpdatePlayerActions(ctx context.Context, gameID string, playerID string, actions []model.PlayerAction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.Actions = actions

	return nil
}

// UpdateForcedFirstAction updates player forced first action
func (r *RepositoryImpl) UpdateForcedFirstAction(ctx context.Context, gameID string, playerID string, action *model.ForcedFirstAction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.ForcedFirstAction = action

	return nil
}
