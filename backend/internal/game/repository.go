package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/features/turn"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GameRepository provides clean CRUD operations and granular updates for games
type Repository interface {
	// Basic CRUD operations
	Create(ctx context.Context, settings model.GameSettings) (model.Game, error)
	GetByID(ctx context.Context, gameID string) (model.Game, error)
	Delete(ctx context.Context, gameID string) error
	List(ctx context.Context, status string) ([]model.Game, error)

	// Granular update methods for specific fields
	UpdateStatus(ctx context.Context, gameID string, status model.GameStatus) error
	UpdatePhase(ctx context.Context, gameID string, phase model.GamePhase) error
	UpdateGlobalParameters(ctx context.Context, gameID string, params model.GlobalParameters) error
	UpdateCurrentTurn(ctx context.Context, gameID string, playerID *string) error
	UpdateBoard(ctx context.Context, gameID string, board model.Board) error

	SetCurrentPlayer(ctx context.Context, gameID string, playerID string) error
	SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error
	AddPlayerID(ctx context.Context, gameID string, playerID string) error
	RemovePlayerID(ctx context.Context, gameID string, playerID string) error
	SetHostPlayer(ctx context.Context, gameID string, playerID string) error
	UpdateGeneration(ctx context.Context, gameID string, generation int) error

	// Tile operations
	UpdateTileOccupancy(ctx context.Context, gameID string, coord model.HexPosition, occupant *model.TileOccupant, ownerID *string) error
	UpdateTemperature(ctx context.Context, gameID string, temperature int) error
	UpdateOxygen(ctx context.Context, gameID string, oxygen int) error
	UpdateOceans(ctx context.Context, gameID string, oceans int) error
}

// GameRepositoryImpl implements GameRepository with in-memory storage
type RepositoryImpl struct {
	games map[string]*model.Game
	mutex sync.RWMutex
	eventBus *events.EventBusImpl

	// Feature repositories (scoped by gameID)
	parametersRepos map[string]parameters.Repository
	boardRepos      map[string]tiles.BoardRepository
	turnOrderRepos  map[string]turn.TurnOrderRepository
}

// NewRepository creates a new game repository
func NewRepository(eventBus *events.EventBusImpl) Repository {
	return &RepositoryImpl{
		games:           make(map[string]*model.Game),
		eventBus:        eventBus,
		parametersRepos: make(map[string]parameters.Repository),
		boardRepos:      make(map[string]tiles.BoardRepository),
		turnOrderRepos:  make(map[string]turn.TurnOrderRepository),
	}
}

// Create creates a new game with the given settings
func (r *RepositoryImpl) Create(ctx context.Context, settings model.GameSettings) (model.Game, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.Get()
	log.Debug("Creating new game")

	// Generate unique game ID
	gameID := uuid.New().String()

	// Create the game metadata (without feature data)
	game := model.NewGame(gameID, settings)

	// Create feature repositories with initial state
	initialParams := parameters.GlobalParameters{
		Temperature: -30,
		Oxygen:      0,
		Oceans:      0,
	}
	parametersRepo, err := parameters.NewRepository(gameID, initialParams, r.eventBus)
	if err != nil {
		return model.Game{}, fmt.Errorf("failed to create parameters repository: %w", err)
	}
	r.parametersRepos[gameID] = parametersRepo

	// Create initial board (42 tiles for standard Mars board)
	initialBoard := tiles.NewStandardBoard()
	boardRepo := tiles.NewBoardRepository(gameID, initialBoard, r.eventBus)
	r.boardRepos[gameID] = boardRepo

	// Create turn order repository (empty player list initially)
	turnOrderRepo := turn.NewTurnOrderRepository([]string{}, nil)
	r.turnOrderRepos[gameID] = turnOrderRepo

	// Create feature services and inject into game
	game.ParametersService = parameters.NewService(parametersRepo)
	game.BoardService = tiles.NewBoardService(boardRepo)
	game.TurnOrderService = turn.NewTurnOrderService(turnOrderRepo)

	// Store game in repository
	r.games[gameID] = game

	log.Debug("Game created with feature services", zap.String("game_id", gameID))

	// Return a copy with services injected
	return r.injectServices(gameID, game), nil
}

// injectServices creates a copy of the game with feature services injected
func (r *RepositoryImpl) injectServices(gameID string, game *model.Game) model.Game {
	gameCopy := *game
	gameCopy.PlayerIDs = make([]string, len(game.PlayerIDs))
	copy(gameCopy.PlayerIDs, game.PlayerIDs)

	// Inject feature services
	if parametersRepo, exists := r.parametersRepos[gameID]; exists {
		gameCopy.ParametersService = parameters.NewService(parametersRepo)
	}
	if boardRepo, exists := r.boardRepos[gameID]; exists {
		gameCopy.BoardService = tiles.NewBoardService(boardRepo)
	}
	if turnOrderRepo, exists := r.turnOrderRepos[gameID]; exists {
		gameCopy.TurnOrderService = turn.NewTurnOrderService(turnOrderRepo)
	}

	return gameCopy
}

// GetByID retrieves a game by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, gameID string) (model.Game, error) {
	if gameID == "" {
		return model.Game{}, fmt.Errorf("game ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	game, exists := r.games[gameID]
	if !exists {
		return model.Game{}, &model.NotFoundError{Resource: "game", ID: gameID}
	}

	// Return a copy with services injected
	return r.injectServices(gameID, game), nil
}

// Delete removes a game from the repository
func (r *RepositoryImpl) Delete(ctx context.Context, gameID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	if gameID == "" {
		return fmt.Errorf("game ID cannot be empty")
	}

	_, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	delete(r.games, gameID)

	log.Info("Game deleted")

	return nil
}

// List returns all games, optionally filtered by status
func (r *RepositoryImpl) List(ctx context.Context, status string) ([]model.Game, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	games := make([]model.Game, 0)

	for gameID, game := range r.games {
		if status == "" || string(game.Status) == status {
			// Return a copy with services injected
			games = append(games, r.injectServices(gameID, game))
		}
	}

	return games, nil
}

// UpdateStatus updates a game's status
func (r *RepositoryImpl) UpdateStatus(ctx context.Context, gameID string, status model.GameStatus) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	oldStatus := game.Status
	game.Status = status
	game.UpdatedAt = time.Now()

	log.Info("Game status updated", zap.String("old_status", string(oldStatus)), zap.String("new_status", string(status)))

	return nil
}

// UpdatePhase updates a game's current phase
func (r *RepositoryImpl) UpdatePhase(ctx context.Context, gameID string, phase model.GamePhase) error {
	// Store old phase value before locking for event publishing
	var oldPhase model.GamePhase
	var shouldPublishEvent bool

	func() {
		r.mutex.Lock()
		defer r.mutex.Unlock()

		log := logger.WithGameContext(gameID, "")

		game, exists := r.games[gameID]
		if !exists {
			return
		}

		oldPhase = game.CurrentPhase
		game.CurrentPhase = phase
		game.UpdatedAt = time.Now()

		log.Info("Game phase updated", zap.String("old_phase", string(oldPhase)), zap.String("new_phase", string(phase)))

		shouldPublishEvent = r.eventBus != nil && oldPhase != phase
	}()

	// Publish event AFTER releasing the lock to avoid deadlock
	if shouldPublishEvent {
		events.Publish(r.eventBus, events.GamePhaseChangedEvent{
			GameID:    gameID,
			OldPhase:  string(oldPhase),
			NewPhase:  string(phase),
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateGlobalParameters updates global parameters for a game
func (r *RepositoryImpl) UpdateGlobalParameters(ctx context.Context, gameID string, params model.GlobalParameters) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	// Validate parameters
	if params.Temperature < -30 || params.Temperature > 8 {
		return fmt.Errorf("temperature must be between -30 and 8, got %d", params.Temperature)
	}
	if params.Oxygen < 0 || params.Oxygen > 14 {
		return fmt.Errorf("oxygen must be between 0 and 14, got %d", params.Oxygen)
	}
	if params.Oceans < 0 || params.Oceans > 9 {
		return fmt.Errorf("oceans must be between 0 and 9, got %d", params.Oceans)
	}

	// Verify game exists
	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Update via feature repository
	parametersRepo, exists := r.parametersRepos[gameID]
	if !exists {
		return fmt.Errorf("parameters repository not found for game %s", gameID)
	}

	// Set all parameters at once
	if err := parametersRepo.SetTemperature(ctx, params.Temperature); err != nil {
		return fmt.Errorf("failed to set temperature: %w", err)
	}
	if err := parametersRepo.SetOxygen(ctx, params.Oxygen); err != nil {
		return fmt.Errorf("failed to set oxygen: %w", err)
	}
	if err := parametersRepo.SetOceans(ctx, params.Oceans); err != nil {
		return fmt.Errorf("failed to set oceans: %w", err)
	}

	game.UpdatedAt = time.Now()

	log.Info("Global parameters updated",
		zap.Int("temperature", params.Temperature),
		zap.Int("oxygen", params.Oxygen),
		zap.Int("oceans", params.Oceans),
	)

	return nil
}

// UpdateCurrentTurn updates the current turn during ongoing gameplay (turn progression)
// Use this when advancing turns during normal game flow (e.g., after a player skips their turn)
func (r *RepositoryImpl) UpdateCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	return r.SetCurrentTurn(ctx, gameID, playerID)
}

// SetCurrentPlayer sets the current active player
func (r *RepositoryImpl) SetCurrentPlayer(ctx context.Context, gameID string, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	oldPlayerID := game.ViewingPlayerID
	game.ViewingPlayerID = playerID
	game.UpdatedAt = time.Now()

	log.Info("Current player updated", zap.String("old_player", oldPlayerID), zap.String("new_player", playerID))

	return nil
}

// SetCurrentTurn sets the initial current turn at game setup (who starts the game)
// Use this when initially setting who begins the game or when resetting turns for new phases
func (r *RepositoryImpl) SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	// Verify game exists
	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Update via feature repository
	turnOrderRepo, exists := r.turnOrderRepos[gameID]
	if !exists {
		return fmt.Errorf("turn order repository not found for game %s", gameID)
	}

	// Get old turn for logging
	oldTurn, _ := turnOrderRepo.GetCurrentTurn(ctx)
	var oldTurnPlayer string
	if oldTurn != nil {
		oldTurnPlayer = *oldTurn
	} else {
		oldTurnPlayer = "none"
	}

	// Set new turn
	if err := turnOrderRepo.SetCurrentTurn(ctx, playerID); err != nil {
		return fmt.Errorf("failed to set current turn: %w", err)
	}

	game.UpdatedAt = time.Now()

	var newTurnPlayer string
	if playerID != nil {
		newTurnPlayer = *playerID
		log.Info("Current turn updated", zap.String("old_turn", oldTurnPlayer), zap.String("new_turn", newTurnPlayer))
	} else {
		newTurnPlayer = "none"
		log.Info("Current turn cleared", zap.String("old_turn", oldTurnPlayer))
	}

	return nil
}

// AddPlayerID adds a player ID to the game
func (r *RepositoryImpl) AddPlayerID(ctx context.Context, gameID string, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Check if player already exists
	for _, existingPlayerID := range game.PlayerIDs {
		if existingPlayerID == playerID {
			return fmt.Errorf("player %s already exists in game %s", playerID, gameID)
		}
	}

	game.PlayerIDs = append(game.PlayerIDs, playerID)
	game.UpdatedAt = time.Now()

	// Update turn order repository with new player list
	if turnOrderRepo, exists := r.turnOrderRepos[gameID]; exists {
		if err := turnOrderRepo.SetPlayerOrder(ctx, game.PlayerIDs); err != nil {
			log.Error("Failed to update player order in turn repository", zap.Error(err))
		}
	}

	// Set as host if first player
	if len(game.PlayerIDs) == 1 {
		game.HostPlayerID = playerID
		log.Info("Player added and set as host")
	} else {
		log.Info("Player added to game")
	}

	return nil
}

// RemovePlayerID removes a player ID from the game
func (r *RepositoryImpl) RemovePlayerID(ctx context.Context, gameID string, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Find and remove player ID
	for i, existingPlayerID := range game.PlayerIDs {
		if existingPlayerID == playerID {
			game.PlayerIDs = append(game.PlayerIDs[:i], game.PlayerIDs[i+1:]...)
			game.UpdatedAt = time.Now()

			// Update turn order repository with new player list
			if turnOrderRepo, exists := r.turnOrderRepos[gameID]; exists {
				if err := turnOrderRepo.SetPlayerOrder(ctx, game.PlayerIDs); err != nil {
					log.Error("Failed to update player order in turn repository", zap.Error(err))
				}
			}

			// Clear host if they were the host
			if game.HostPlayerID == playerID {
				if len(game.PlayerIDs) > 0 {
					game.HostPlayerID = game.PlayerIDs[0] // Set first remaining player as host
					log.Info("Player removed and host transferred", zap.String("new_host", game.HostPlayerID))
				} else {
					game.HostPlayerID = ""
					log.Info("Player removed and no host remaining")
				}
			} else {
				log.Info("Player removed from game")
			}

			return nil
		}
	}

	return fmt.Errorf("player %s not found in game %s", playerID, gameID)
}

// SetHostPlayer sets the host player for the game
func (r *RepositoryImpl) SetHostPlayer(ctx context.Context, gameID string, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Verify player exists in game
	playerExists := false
	for _, existingPlayerID := range game.PlayerIDs {
		if existingPlayerID == playerID {
			playerExists = true
			break
		}
	}

	if !playerExists {
		return fmt.Errorf("player %s not found in game %s", playerID, gameID)
	}

	game.HostPlayerID = playerID
	game.UpdatedAt = time.Now()

	log.Info("Host player updated")

	return nil
}

// UpdateGeneration updates the game generation
func (r *RepositoryImpl) UpdateGeneration(ctx context.Context, gameID string, generation int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	game.Generation = generation
	game.UpdatedAt = time.Now()

	log.Info("Generation updated", zap.Int("generation", generation))

	return nil
}

// UpdateBoard updates the board state for a game
// DEPRECATED: This method is kept for compatibility but should use individual tile operations
func (r *RepositoryImpl) UpdateBoard(ctx context.Context, gameID string, board model.Board) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	// Verify game exists
	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Note: Board is now managed by feature repository
	// This method is kept for compatibility but individual tile operations should be used
	game.UpdatedAt = time.Now()

	log.Debug("Board update requested (managed by feature repository)", zap.Int("tiles", len(board.Tiles)))

	return nil
}


// UpdateTileOccupancy updates a tile's occupancy and ownership
func (r *RepositoryImpl) UpdateTileOccupancy(ctx context.Context, gameID string, coord model.HexPosition, occupant *model.TileOccupant, ownerID *string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	// Verify game exists
	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Update via feature repository
	boardRepo, exists := r.boardRepos[gameID]
	if !exists {
		return fmt.Errorf("board repository not found for game %s", gameID)
	}

	// Convert model types to tiles types
	tilesCoord := tiles.HexPosition{Q: coord.Q, R: coord.R, S: coord.S}
	var tilesOccupant tiles.TileOccupant
	if occupant != nil {
		tilesOccupant = tiles.TileOccupant{
			Type: tiles.ResourceType(occupant.Type),
			Tags: occupant.Tags,
		}
	}

	// Place tile via repository (this publishes TilePlacedEvent)
	if err := boardRepo.PlaceTile(ctx, tilesCoord, tilesOccupant, ownerID); err != nil {
		return fmt.Errorf("failed to place tile: %w", err)
	}

	game.UpdatedAt = time.Now()

	occupantType := "empty"
	ownerName := "none"
	if occupant != nil {
		occupantType = string(occupant.Type)
	}
	if ownerID != nil {
		ownerName = *ownerID
	}

	log.Info("Tile occupancy updated",
		zap.Int("q", coord.Q),
		zap.Int("r", coord.R),
		zap.Int("s", coord.S),
		zap.String("occupant", occupantType),
		zap.String("owner", ownerName))

	return nil
}

// UpdateTemperature updates the global temperature parameter
func (r *RepositoryImpl) UpdateTemperature(ctx context.Context, gameID string, temperature int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	// Verify game exists
	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Update via feature repository
	parametersRepo, exists := r.parametersRepos[gameID]
	if !exists {
		return fmt.Errorf("parameters repository not found for game %s", gameID)
	}

	// Get old temperature for logging
	oldParams, _ := parametersRepo.Get(ctx)
	oldTemp := oldParams.Temperature

	// Set temperature (this publishes TemperatureChangedEvent)
	if err := parametersRepo.SetTemperature(ctx, temperature); err != nil {
		return fmt.Errorf("failed to set temperature: %w", err)
	}

	game.UpdatedAt = time.Now()

	log.Info("Temperature updated",
		zap.Int("old_temperature", oldTemp),
		zap.Int("new_temperature", temperature))

	return nil
}

// UpdateOxygen updates the global oxygen parameter
func (r *RepositoryImpl) UpdateOxygen(ctx context.Context, gameID string, oxygen int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	// Verify game exists
	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Update via feature repository
	parametersRepo, exists := r.parametersRepos[gameID]
	if !exists {
		return fmt.Errorf("parameters repository not found for game %s", gameID)
	}

	// Get old oxygen for logging
	oldParams, _ := parametersRepo.Get(ctx)
	oldOxygen := oldParams.Oxygen

	// Set oxygen (this publishes OxygenChangedEvent)
	if err := parametersRepo.SetOxygen(ctx, oxygen); err != nil {
		return fmt.Errorf("failed to set oxygen: %w", err)
	}

	game.UpdatedAt = time.Now()

	log.Info("Oxygen updated",
		zap.Int("old_oxygen", oldOxygen),
		zap.Int("new_oxygen", oxygen))

	return nil
}

// UpdateOceans updates the global oceans parameter
func (r *RepositoryImpl) UpdateOceans(ctx context.Context, gameID string, oceans int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	// Verify game exists
	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Update via feature repository
	parametersRepo, exists := r.parametersRepos[gameID]
	if !exists {
		return fmt.Errorf("parameters repository not found for game %s", gameID)
	}

	// Get old oceans for logging
	oldParams, _ := parametersRepo.Get(ctx)
	oldOceans := oldParams.Oceans

	// Set oceans (this publishes OceansChangedEvent)
	if err := parametersRepo.SetOceans(ctx, oceans); err != nil {
		return fmt.Errorf("failed to set oceans: %w", err)
	}

	game.UpdatedAt = time.Now()

	log.Info("Oceans updated",
		zap.Int("old_oceans", oldOceans),
		zap.Int("new_oceans", oceans))

	return nil
}

// Clear removes all games from the repository
func (r *RepositoryImpl) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.games = make(map[string]*model.Game)
	r.parametersRepos = make(map[string]parameters.Repository)
	r.boardRepos = make(map[string]tiles.BoardRepository)
	r.turnOrderRepos = make(map[string]turn.TurnOrderRepository)
}
