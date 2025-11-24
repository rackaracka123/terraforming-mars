package player

import (
	"context"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/types"
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
	UpdateResources(ctx context.Context, gameID string, playerID string, resources types.Resources) error

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
	UpdateProduction(ctx context.Context, gameID string, playerID string, production types.Production) error

	// UpdateSelectStartingCardsPhase updates the starting cards selection phase
	UpdateSelectStartingCardsPhase(ctx context.Context, gameID string, playerID string, phase *SelectStartingCardsPhase) error

	// UpdateProductionPhase updates the production phase
	UpdateProductionPhase(ctx context.Context, gameID string, playerID string, phase *types.ProductionPhase) error

	// UpdateCorporation updates the player's corporation with full card data
	UpdateCorporation(ctx context.Context, gameID string, playerID string, corporation types.Card) error

	// UpdatePaymentSubstitutes updates player payment substitutes
	UpdatePaymentSubstitutes(ctx context.Context, gameID string, playerID string, substitutes []types.PaymentSubstitute) error

	// UpdatePlayerActions updates player actions
	UpdatePlayerActions(ctx context.Context, gameID string, playerID string, actions []types.PlayerAction) error

	// UpdateForcedFirstAction updates player forced first action
	UpdateForcedFirstAction(ctx context.Context, gameID string, playerID string, action *types.ForcedFirstAction) error

	// UpdateRequirementModifiers updates player requirement modifiers
	UpdateRequirementModifiers(ctx context.Context, gameID string, playerID string, modifiers []types.RequirementModifier) error

	// UpdatePlayerEffects updates player active effects
	UpdatePlayerEffects(ctx context.Context, gameID string, playerID string, effects []types.PlayerEffect) error

	// UpdateTerraformRating updates player terraform rating
	UpdateTerraformRating(ctx context.Context, gameID string, playerID string, rating int) error

	// UpdateVictoryPoints updates player victory points
	UpdateVictoryPoints(ctx context.Context, gameID string, playerID string, victoryPoints int) error

	// CreateTileQueue creates a tile placement queue for the player
	CreateTileQueue(ctx context.Context, gameID string, playerID string, cardID string, tileTypes []string) error

	// GetPendingTileSelectionQueue retrieves the pending tile selection queue for a player
	GetPendingTileSelectionQueue(ctx context.Context, gameID string, playerID string) (*types.PendingTileSelectionQueue, error)

	// ProcessNextTileInQueue pops the next tile type from the queue and returns it
	ProcessNextTileInQueue(ctx context.Context, gameID string, playerID string) (string, error)

	// UpdatePendingTileSelection updates the pending tile selection for a player
	UpdatePendingTileSelection(ctx context.Context, gameID string, playerID string, selection *types.PendingTileSelection) error

	// ClearPendingTileSelection clears the pending tile selection for a player
	ClearPendingTileSelection(ctx context.Context, gameID string, playerID string) error

	// UpdatePendingTileSelectionQueue updates the pending tile selection queue
	UpdatePendingTileSelectionQueue(ctx context.Context, gameID string, playerID string, queue *types.PendingTileSelectionQueue) error

	// ClearPendingTileSelectionQueue clears the pending tile selection queue
	ClearPendingTileSelectionQueue(ctx context.Context, gameID string, playerID string) error

	// UpdatePendingCardDrawSelection updates player pending card draw selection
	UpdatePendingCardDrawSelection(ctx context.Context, gameID string, playerID string, selection *types.PendingCardDrawSelection) error

	// UpdateResourceStorage updates player resource storage
	UpdateResourceStorage(ctx context.Context, gameID string, playerID string, storage map[string]int) error

	// RemoveCardFromHand removes a card from the player's hand
	RemoveCardFromHand(ctx context.Context, gameID string, playerID string, cardID string) error

	// UpdatePassed updates player passed status for generation
	UpdatePassed(ctx context.Context, gameID string, playerID string, passed bool) error

	// UpdateAvailableActions updates player available actions count
	UpdateAvailableActions(ctx context.Context, gameID string, playerID string, actions int) error

	// UpdatePendingCardSelection updates player pending card selection
	UpdatePendingCardSelection(ctx context.Context, gameID string, playerID string, selection *PendingCardSelection) error

	// ClearPendingCardSelection clears the pending card selection
	ClearPendingCardSelection(ctx context.Context, gameID string, playerID string) error
}

// RepositoryImpl implements the Repository interface using facade pattern
// Delegates to focused sub-repositories for each domain concern
type RepositoryImpl struct {
	core        *PlayerCoreRepository
	resource    *PlayerResourceRepository
	hand        *PlayerHandRepository
	selection   *PlayerSelectionRepository
	corporation *PlayerCorporationRepository
	turn        *PlayerTurnRepository
	effect      *PlayerEffectRepository
	tileQueue   *PlayerTileQueueRepository
}

// NewRepository creates a new player repository facade
func NewRepository(eventBus *events.EventBusImpl) Repository {
	// Create shared storage
	storage := NewPlayerStorage()

	// Create sub-repositories
	coreRepo := NewPlayerCoreRepository(storage)
	resourceRepo := NewPlayerResourceRepository(storage, eventBus)
	handRepo := NewPlayerHandRepository(storage)
	selectionRepo := NewPlayerSelectionRepository(storage)
	corporationRepo := NewPlayerCorporationRepository(storage)
	turnRepo := NewPlayerTurnRepository(storage)
	effectRepo := NewPlayerEffectRepository(storage)
	tileQueueRepo := NewPlayerTileQueueRepository(storage, eventBus)

	return &RepositoryImpl{
		core:        coreRepo,
		resource:    resourceRepo,
		hand:        handRepo,
		selection:   selectionRepo,
		corporation: corporationRepo,
		turn:        turnRepo,
		effect:      effectRepo,
		tileQueue:   tileQueueRepo,
	}
}

// ============================================================================
// CORE OPERATIONS (delegated to core repository)
// ============================================================================

// Create creates a new player in a game
func (r *RepositoryImpl) Create(ctx context.Context, gameID string, player *Player) error {
	return r.core.Create(ctx, gameID, player)
}

// GetByID retrieves a player by ID from a specific game
func (r *RepositoryImpl) GetByID(ctx context.Context, gameID string, playerID string) (*Player, error) {
	return r.core.GetByID(ctx, gameID, playerID)
}

// ListByGameID retrieves all players in a game
func (r *RepositoryImpl) ListByGameID(ctx context.Context, gameID string) ([]*Player, error) {
	return r.core.ListByGameID(ctx, gameID)
}

// ============================================================================
// RESOURCE OPERATIONS (delegated to resource repository)
// ============================================================================
