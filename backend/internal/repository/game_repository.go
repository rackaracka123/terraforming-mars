package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GameRepository provides clean CRUD operations and granular updates for games
type GameRepository interface {
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
type GameRepositoryImpl struct {
	games    map[string]*model.Game
	mutex    sync.RWMutex
	eventBus *events.EventBusImpl
}

// NewGameRepository creates a new game repository
func NewGameRepository(eventBus *events.EventBusImpl) GameRepository {
	return &GameRepositoryImpl{
		games:    make(map[string]*model.Game),
		eventBus: eventBus,
	}
}

// Create creates a new game with the given settings
func (r *GameRepositoryImpl) Create(ctx context.Context, settings model.GameSettings) (model.Game, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.Get()
	log.Debug("Creating new game")

	// Generate unique game ID
	gameID := uuid.New().String()

	// Create the game
	game := model.NewGame(gameID, settings)

	// Store in repository
	r.games[gameID] = game

	log.Debug("Game created", zap.String("game_id", gameID))

	return *game, nil
}

// GetByID retrieves a game by ID
func (r *GameRepositoryImpl) GetByID(ctx context.Context, gameID string) (model.Game, error) {
	if gameID == "" {
		return model.Game{}, fmt.Errorf("game ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	game, exists := r.games[gameID]
	if !exists {
		return model.Game{}, &model.NotFoundError{Resource: "game", ID: gameID}
	}

	// Return a copy to prevent external mutation
	gameCopy := *game
	gameCopy.PlayerIDs = make([]string, len(game.PlayerIDs))
	copy(gameCopy.PlayerIDs, game.PlayerIDs)

	return gameCopy, nil
}

// Delete removes a game from the repository
func (r *GameRepositoryImpl) Delete(ctx context.Context, gameID string) error {
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
func (r *GameRepositoryImpl) List(ctx context.Context, status string) ([]model.Game, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	games := make([]model.Game, 0)

	for _, game := range r.games {
		if status == "" || string(game.Status) == status {
			// Return a copy to prevent external mutation
			gameCopy := *game
			gameCopy.PlayerIDs = make([]string, len(game.PlayerIDs))
			copy(gameCopy.PlayerIDs, game.PlayerIDs)
			games = append(games, gameCopy)
		}
	}

	return games, nil
}

// UpdateStatus updates a game's status
func (r *GameRepositoryImpl) UpdateStatus(ctx context.Context, gameID string, status model.GameStatus) error {
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
func (r *GameRepositoryImpl) UpdatePhase(ctx context.Context, gameID string, phase model.GamePhase) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	oldPhase := game.CurrentPhase
	game.CurrentPhase = phase
	game.UpdatedAt = time.Now()

	log.Info("Game phase updated", zap.String("old_phase", string(oldPhase)), zap.String("new_phase", string(phase)))

	return nil
}

// UpdateGlobalParameters updates global parameters for a game
func (r *GameRepositoryImpl) UpdateGlobalParameters(ctx context.Context, gameID string, params model.GlobalParameters) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	if err := r.validateGlobalParameters(&params); err != nil {
		log.Error("Invalid global parameters", zap.Error(err))
		return fmt.Errorf("invalid parameters: %w", err)
	}

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	game.GlobalParameters = params
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
func (r *GameRepositoryImpl) UpdateCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	return r.SetCurrentTurn(ctx, gameID, playerID)
}

// SetCurrentPlayer sets the current active player
func (r *GameRepositoryImpl) SetCurrentPlayer(ctx context.Context, gameID string, playerID string) error {
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
func (r *GameRepositoryImpl) SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	var oldTurnPlayer string
	if game.CurrentTurn != nil {
		oldTurnPlayer = *game.CurrentTurn
	} else {
		oldTurnPlayer = "none"
	}

	game.CurrentTurn = playerID
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
func (r *GameRepositoryImpl) AddPlayerID(ctx context.Context, gameID string, playerID string) error {
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
func (r *GameRepositoryImpl) RemovePlayerID(ctx context.Context, gameID string, playerID string) error {
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
func (r *GameRepositoryImpl) SetHostPlayer(ctx context.Context, gameID string, playerID string) error {
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
func (r *GameRepositoryImpl) UpdateGeneration(ctx context.Context, gameID string, generation int) error {
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
func (r *GameRepositoryImpl) UpdateBoard(ctx context.Context, gameID string, board model.Board) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	game.Board = board
	game.UpdatedAt = time.Now()

	log.Debug("Board updated", zap.Int("tiles", len(board.Tiles)))

	return nil
}

// validateGlobalParameters ensures parameters are within valid game ranges
func (r *GameRepositoryImpl) validateGlobalParameters(params *model.GlobalParameters) error {
	if params.Temperature < -30 || params.Temperature > 8 {
		return fmt.Errorf("temperature must be between -30 and 8, got %d", params.Temperature)
	}

	if params.Oxygen < 0 || params.Oxygen > 14 {
		return fmt.Errorf("oxygen must be between 0 and 14, got %d", params.Oxygen)
	}

	if params.Oceans < 0 || params.Oceans > 9 {
		return fmt.Errorf("oceans must be between 0 and 9, got %d", params.Oceans)
	}

	return nil
}

// UpdateTileOccupancy updates a tile's occupancy and ownership
func (r *GameRepositoryImpl) UpdateTileOccupancy(ctx context.Context, gameID string, coord model.HexPosition, occupant *model.TileOccupant, ownerID *string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Find the tile to update
	var targetTile *model.Tile
	for i := range game.Board.Tiles {
		tile := &game.Board.Tiles[i]
		if tile.Coordinates.Q == coord.Q && tile.Coordinates.R == coord.R && tile.Coordinates.S == coord.S {
			targetTile = tile
			break
		}
	}

	if targetTile == nil {
		return fmt.Errorf("tile not found at coordinate %d,%d,%d", coord.Q, coord.R, coord.S)
	}

	// Update occupancy and ownership
	targetTile.OccupiedBy = occupant
	targetTile.OwnerID = ownerID

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

	// Publish tile placed event when a tile occupant is added
	if r.eventBus != nil && occupant != nil && ownerID != nil {
		events.Publish(r.eventBus, TilePlacedEvent{
			GameID:    gameID,
			PlayerID:  *ownerID,
			TileType:  string(occupant.Type),
			Q:         coord.Q,
			R:         coord.R,
			S:         coord.S,
			Timestamp: time.Now(),
		})
		log.Debug("🎆 TilePlacedEvent published",
			zap.String("tile_type", string(occupant.Type)),
			zap.String("player_id", *ownerID))
	}

	return nil
}

// UpdateTemperature updates the global temperature parameter
func (r *GameRepositoryImpl) UpdateTemperature(ctx context.Context, gameID string, temperature int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Validate temperature range
	if temperature < -30 || temperature > 8 {
		return fmt.Errorf("temperature must be between -30 and 8, got %d", temperature)
	}

	oldTemp := game.GlobalParameters.Temperature
	game.GlobalParameters.Temperature = temperature
	game.UpdatedAt = time.Now()

	log.Info("Temperature updated",
		zap.Int("old_temperature", oldTemp),
		zap.Int("new_temperature", temperature))

	// Publish temperature changed event
	if r.eventBus != nil && oldTemp != temperature {
		events.Publish(r.eventBus, TemperatureChangedEvent{
			GameID:    gameID,
			OldValue:  oldTemp,
			NewValue:  temperature,
			ChangedBy: "", // Will be populated by service layer in future enhancement
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateOxygen updates the global oxygen parameter
func (r *GameRepositoryImpl) UpdateOxygen(ctx context.Context, gameID string, oxygen int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Validate oxygen range
	if oxygen < 0 || oxygen > 14 {
		return fmt.Errorf("oxygen must be between 0 and 14, got %d", oxygen)
	}

	oldOxygen := game.GlobalParameters.Oxygen
	game.GlobalParameters.Oxygen = oxygen
	game.UpdatedAt = time.Now()

	log.Info("Oxygen updated",
		zap.Int("old_oxygen", oldOxygen),
		zap.Int("new_oxygen", oxygen))

	// Publish oxygen changed event
	if r.eventBus != nil && oldOxygen != oxygen {
		events.Publish(r.eventBus, OxygenChangedEvent{
			GameID:    gameID,
			OldValue:  oldOxygen,
			NewValue:  oxygen,
			ChangedBy: "", // Will be populated by service layer in future enhancement
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateOceans updates the global oceans parameter
func (r *GameRepositoryImpl) UpdateOceans(ctx context.Context, gameID string, oceans int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Validate oceans range
	if oceans < 0 || oceans > 9 {
		return fmt.Errorf("oceans must be between 0 and 9, got %d", oceans)
	}

	oldOceans := game.GlobalParameters.Oceans
	game.GlobalParameters.Oceans = oceans
	game.UpdatedAt = time.Now()

	log.Info("Oceans updated",
		zap.Int("old_oceans", oldOceans),
		zap.Int("new_oceans", oceans))

	// Publish oceans changed event
	if r.eventBus != nil && oldOceans != oceans {
		events.Publish(r.eventBus, OceansChangedEvent{
			GameID:    gameID,
			OldValue:  oldOceans,
			NewValue:  oceans,
			ChangedBy: "", // Will be populated by service layer in future enhancement
			Timestamp: time.Now(),
		})
	}

	return nil
}
