package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GameService handles all game lifecycle operations including creation, player management, and state updates
type GameService interface {
	// Create a new game with specified settings
	CreateGame(ctx context.Context, settings model.GameSettings) (model.Game, error)

	// Get game by ID
	GetGame(ctx context.Context, gameID string) (model.Game, error)

	// Get game prepared for a specific player's perspective
	GetGameForPlayer(ctx context.Context, gameID string, playerID string) (model.Game, error)

	// List games by status
	ListGames(ctx context.Context, status string) ([]model.Game, error)

	// Start a game (transition from status "lobby" to "active")
	StartGame(ctx context.Context, gameID string, playerID string) error

	// Advance game phase after all players complete starting card selection
	AdvanceFromCardSelectionPhase(ctx context.Context, gameID string) error

	// Add player to game (join game flow)
	JoinGame(ctx context.Context, gameID string, playerName string) (model.Game, error)

	// Global parameters methods (merged from GlobalParametersService)
	UpdateGlobalParameters(ctx context.Context, gameID string, newParams model.GlobalParameters) error
	GetGlobalParameters(ctx context.Context, gameID string) (model.GlobalParameters, error)
	IncreaseTemperature(ctx context.Context, gameID string, steps int) error
	IncreaseOxygen(ctx context.Context, gameID string, steps int) error
	PlaceOcean(ctx context.Context, gameID string, count int) error
}

// GameServiceImpl implements GameService interface
type GameServiceImpl struct {
	gameRepo    repository.GameRepository
	playerRepo  repository.PlayerRepository
	cardService *CardServiceImpl // Use concrete type to access StorePlayerCardOptions
	eventBus    events.EventBus
}

// NewGameService creates a new GameService instance
func NewGameService(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	cardService *CardServiceImpl,
	eventBus events.EventBus,
) GameService {
	return &GameServiceImpl{
		gameRepo:    gameRepo,
		playerRepo:  playerRepo,
		cardService: cardService,
		eventBus:    eventBus,
	}
}

// CreateGame creates a new game with specified settings
func (s *GameServiceImpl) CreateGame(ctx context.Context, settings model.GameSettings) (model.Game, error) {
	log := logger.WithContext()

	// Validate settings
	if err := s.validateGameSettings(settings); err != nil {
		log.Error("Invalid game settings", zap.Error(err))
		return model.Game{}, fmt.Errorf("invalid game settings: %w", err)
	}

	log.Debug("Creating game via GameService")

	game, err := s.gameRepo.Create(ctx, settings)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to create game: %w", err)
	}

	log.Info("Game created via GameService", zap.String("game_id", game.ID))
	return game, nil
}

// GetGame retrieves a game by ID
func (s *GameServiceImpl) GetGame(ctx context.Context, gameID string) (model.Game, error) {
	return s.gameRepo.Get(ctx, gameID)
}

// GetGameForPlayer gets a game prepared for a specific player's perspective
func (s *GameServiceImpl) GetGameForPlayer(ctx context.Context, gameID string, playerID string) (model.Game, error) {
	// Get the full game data
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		return model.Game{}, err
	}

	// Create a copy of the game to modify
	gameCopy := game

	// Find the viewing player and set as CurrentPlayer
	gameCopy.CurrentPlayer = nil
	for i := range game.Players {
		if game.Players[i].ID == playerID {
			gameCopy.CurrentPlayer = game.Players[i].DeepCopy()
			break
		}
	}

	// Populate OtherPlayers with limited data for all players except the viewing player
	gameCopy.OtherPlayers = make([]model.OtherPlayer, 0, len(game.Players)-1)
	for i := range game.Players {
		if game.Players[i].ID != playerID {
			otherPlayer := model.NewOtherPlayerFromPlayer(&game.Players[i])
			if otherPlayer != nil {
				gameCopy.OtherPlayers = append(gameCopy.OtherPlayers, *otherPlayer)
			}
		}
	}

	return gameCopy, nil
}

// ListGames lists games by status
func (s *GameServiceImpl) ListGames(ctx context.Context, status string) ([]model.Game, error) {
	return s.gameRepo.List(ctx, status)
}

func (s *GameServiceImpl) StartGame(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Starting game via GameService")

	// Get current game state
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for start", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Ensure player is host
	if game.HostPlayerID != playerID {
		log.Warn("Non-host player attempted to start game", zap.String("player_id", playerID))
		return fmt.Errorf("only the host can start the game")
	}

	// Validate game can be started
	if game.Status != model.GameStatusLobby {
		log.Warn("Attempted to start game not in lobby state", zap.String("current_status", string(game.Status)))
		return fmt.Errorf("game is not in lobby state")
	}
	if len(game.Players) < 1 {
		log.Warn("Attempted to start game with no players")
		return fmt.Errorf("cannot start game with no players")
	}

	// Transition game status to active and phase to starting card selection
	game.Status = model.GameStatusActive
	game.CurrentPhase = model.GamePhaseStartingCardSelection

	// Update game through repository
	if err := s.gameRepo.Update(ctx, &game); err != nil {
		log.Error("Failed to update game status to active", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	// Distribute starting cards to all players
	if err := s.distributeStartingCards(ctx, gameID, game.Players); err != nil {
		log.Error("Failed to distribute starting cards", zap.Error(err))
		return fmt.Errorf("failed to distribute starting cards: %w", err)
	}

	log.Info("Game started", zap.String("game_id", gameID))
	return nil
}

// AddPlayerToGame adds a player to the game (business logic from Game model)
func (s *GameServiceImpl) AddPlayerToGame(game *model.Game, player model.Player) bool {
	if len(game.Players) >= game.Settings.MaxPlayers {
		return false
	}

	game.Players = append(game.Players, player)
	game.UpdatedAt = time.Now()

	return true
}

// GetPlayerFromGame returns a player by ID (business logic from Game model)
func (s *GameServiceImpl) GetPlayerFromGame(game *model.Game, playerID string) (*model.Player, bool) {
	for i := range game.Players {
		if game.Players[i].ID == playerID {
			return &game.Players[i], true
		}
	}
	return nil, false
}

// IsGameFull returns true if the game has reached maximum players (business logic from Game model)
func (s *GameServiceImpl) IsGameFull(game *model.Game) bool {
	return len(game.Players) >= game.Settings.MaxPlayers
}

// IsHost returns true if the given player ID is the host of the game (business logic from Game model)
func (s *GameServiceImpl) IsHost(game *model.Game, playerID string) bool {
	return game.HostPlayerID == playerID
}

// JoinGame adds a player to a game using both GameState and Player repositories
func (s *GameServiceImpl) JoinGame(ctx context.Context, gameID string, playerName string) (model.Game, error) {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Player joining game via GameService", zap.String("player_name", playerName))

	// Get the current game state
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for join", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to get game: %w", err)
	}

	// Check if game is joinable
	if game.Status == model.GameStatusCompleted {
		log.Warn("Attempted to join completed game", zap.String("player_name", playerName))
		return model.Game{}, fmt.Errorf("cannot join completed game")
	}

	if s.IsGameFull(&game) {
		log.Warn("Attempted to join full game",
			zap.String("player_name", playerName),
			zap.Int("current_players", len(game.Players)),
		)
		return model.Game{}, fmt.Errorf("game is full")
	}

	// Create new player
	playerID := uuid.New().String()
	player := model.Player{
		ID:   playerID,
		Name: playerName,
		Resources: model.Resources{
			Credits: 40, // Starting credits for standard projects
		},
		Production: model.Production{
			Credits: 1, // Base production
		},
		TerraformRating:  20, // Starting terraform rating
		IsActive:         true,
		PlayedCards:      make([]string, 0),
		Passed:           false, // Player starts active, not passed
		AvailableActions: 2,     // Standard actions per turn in action phase
		VictoryPoints:    0,     // Starting victory points
		MilestoneIcon:    "",    // No milestone achieved initially
		ConnectionStatus: model.ConnectionStatusConnected,
	}

	// Add player through PlayerRepository
	if err := s.playerRepo.AddPlayer(ctx, gameID, player); err != nil {
		log.Error("Failed to add player", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to add player: %w", err)
	}

	// Update game state to include the new player
	if !s.AddPlayerToGame(&game, player) {
		log.Error("Failed to add player to game state")
		return model.Game{}, fmt.Errorf("failed to add player to game")
	}

	// Set the first player as host if no host is set
	if game.HostPlayerID == "" {
		game.HostPlayerID = player.ID
		log.Debug("Player set as host", zap.String("player_id", playerID))
	}

	// Update game through GameStateRepository
	if err := s.gameRepo.Update(ctx, &game); err != nil {
		log.Error("Failed to update game after player join", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to update game: %w", err)
	}

	log.Debug("Player joined game",
		zap.String("player_id", playerID),
		zap.Int("total_players", len(game.Players)),
	)

	return game, nil
}

// validateGameSettings validates game creation settings
func (s *GameServiceImpl) validateGameSettings(settings model.GameSettings) error {
	if settings.MaxPlayers < 1 || settings.MaxPlayers > 5 {
		return fmt.Errorf("max players must be between 1 and 5, got %d", settings.MaxPlayers)
	}

	if settings.Oxygen != nil {
		if *settings.Oxygen < 0 || *settings.Oxygen > 14 {
			return fmt.Errorf("oxygen level must be between 0 and 14, got %d", settings.Oxygen)
		}
	}

	if settings.Temperature != nil {
		if *settings.Temperature < -30 || *settings.Temperature > 8 {
			return fmt.Errorf("temperature must be between -30 and 8, got %d", settings.Temperature)
		}
	}

	if settings.Oceans != nil {
		if *settings.Oceans < 0 || *settings.Oceans > 9 {
			return fmt.Errorf("oceans must be between 0 and 9, got %d", settings.Oceans)
		}
	}
	return nil
}

// distributeStartingCards deals starting card options to all players
func (s *GameServiceImpl) distributeStartingCards(ctx context.Context, gameID string, players []model.Player) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Distributing starting cards to players", zap.Int("player_count", len(players)))

	// Get all available starting cards
	allStartingCards := model.GetStartingCards()
	startingCardIDs := make([]string, len(allStartingCards))
	for i, card := range allStartingCards {
		startingCardIDs[i] = card.ID
	}

	// Create random source for card distribution
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Distribute 4 random cards to each player
	const cardsPerPlayer = 4
	for _, player := range players {
		// Shuffle and select 4 cards
		shuffled := make([]string, len(startingCardIDs))
		copy(shuffled, startingCardIDs)

		// Fisher-Yates shuffle
		for i := len(shuffled) - 1; i > 0; i-- {
			j := rng.Intn(i + 1)
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		}

		cardOptions := shuffled[:cardsPerPlayer]

		log.Debug("Dealing starting cards to player",
			zap.String("player_id", player.ID),
			zap.Strings("cards", cardOptions))

		// Store card options in CardService for validation during selection
		s.cardService.StorePlayerCardOptions(gameID, player.ID, cardOptions)

		// Create and publish event
		event := events.NewPlayerStartingCardOptionsEvent(gameID, player.ID, cardOptions)

		// Publish the event through the event bus
		if s.eventBus != nil {
			if err := s.eventBus.Publish(ctx, event); err != nil {
				log.Warn("Failed to publish starting card options event",
					zap.String("player_id", player.ID),
					zap.Error(err))
			} else {
				log.Debug("Starting card options event published",
					zap.String("player_id", player.ID),
					zap.String("event_type", event.GetType()))
			}
		}
	}

	log.Info("Starting cards distributed to all players", zap.Int("players", len(players)))
	return nil
}

// AdvanceFromCardSelectionPhase advances the game from starting card selection to action phase
func (s *GameServiceImpl) AdvanceFromCardSelectionPhase(ctx context.Context, gameID string) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Advancing game phase from card selection")

	// Get current game state
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for phase advancement", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Validate current phase
	if game.CurrentPhase != model.GamePhaseStartingCardSelection {
		log.Warn("Attempted to advance from card selection phase but game is not in that phase",
			zap.String("current_phase", string(game.CurrentPhase)))
		return fmt.Errorf("game is not in starting card selection phase")
	}

	// Validate game is active
	if game.Status != model.GameStatusActive {
		log.Warn("Attempted to advance phase but game is not active",
			zap.String("current_status", string(game.Status)))
		return fmt.Errorf("game is not active")
	}

	// Advance to action phase
	oldPhase := game.CurrentPhase
	game.CurrentPhase = model.GamePhaseAction

	// Update game through repository
	if err := s.gameRepo.Update(ctx, &game); err != nil {
		log.Error("Failed to update game phase", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	// Clear temporary card selection data
	if s.cardService != nil {
		s.cardService.ClearGameSelectionData(gameID)
	}

	log.Info("Game phase advanced successfully",
		zap.String("previous_phase", string(oldPhase)),
		zap.String("new_phase", string(game.CurrentPhase)))

	return nil
}

// UpdateGlobalParameters updates global terraforming parameters
func (s *GameServiceImpl) UpdateGlobalParameters(ctx context.Context, gameID string, newParams model.GlobalParameters) error {
	log := logger.WithGameContext(gameID, "")

	log.Info("Updating global parameters via GameService",
		zap.Int("temperature", newParams.Temperature),
		zap.Int("oxygen", newParams.Oxygen),
		zap.Int("oceans", newParams.Oceans))

	// Update through GameRepository (now includes global parameters)
	return s.gameRepo.UpdateGlobalParameters(ctx, gameID, &newParams)
}

// GetGlobalParameters gets current global parameters
func (s *GameServiceImpl) GetGlobalParameters(ctx context.Context, gameID string) (model.GlobalParameters, error) {
	return s.gameRepo.GetGlobalParameters(ctx, gameID)
}

// IncreaseTemperature increases temperature by specified steps
func (s *GameServiceImpl) IncreaseTemperature(ctx context.Context, gameID string, steps int) error {
	log := logger.WithGameContext(gameID, "")

	// Get current parameters
	params, err := s.gameRepo.GetGlobalParameters(ctx, gameID)
	if err != nil {
		log.Error("Failed to get global parameters", zap.Error(err))
		return fmt.Errorf("failed to get parameters: %w", err)
	}

	// Calculate new temperature (max +8°C)
	newTemp := params.Temperature + (steps * 2) // Each step = 2°C
	if newTemp > 8 {
		newTemp = 8
	}

	// Update parameters
	updatedParams := params
	updatedParams.Temperature = newTemp

	return s.UpdateGlobalParameters(ctx, gameID, updatedParams)
}

// IncreaseOxygen increases oxygen by specified steps
func (s *GameServiceImpl) IncreaseOxygen(ctx context.Context, gameID string, steps int) error {
	log := logger.WithGameContext(gameID, "")

	// Get current parameters
	params, err := s.gameRepo.GetGlobalParameters(ctx, gameID)
	if err != nil {
		log.Error("Failed to get global parameters", zap.Error(err))
		return fmt.Errorf("failed to get parameters: %w", err)
	}

	// Calculate new oxygen level (max 14%)
	newOxygen := params.Oxygen + steps
	if newOxygen > 14 {
		newOxygen = 14
	}

	// Update parameters
	updatedParams := params
	updatedParams.Oxygen = newOxygen

	return s.UpdateGlobalParameters(ctx, gameID, updatedParams)
}

// PlaceOcean places ocean tiles
func (s *GameServiceImpl) PlaceOcean(ctx context.Context, gameID string, count int) error {
	log := logger.WithGameContext(gameID, "")

	// Get current parameters
	params, err := s.gameRepo.GetGlobalParameters(ctx, gameID)
	if err != nil {
		log.Error("Failed to get global parameters", zap.Error(err))
		return fmt.Errorf("failed to get parameters: %w", err)
	}

	// Calculate new ocean count (max 9 oceans)
	newOceans := params.Oceans + count
	if newOceans > 9 {
		newOceans = 9
	}

	// Update parameters
	updatedParams := params
	updatedParams.Oceans = newOceans

	return s.UpdateGlobalParameters(ctx, gameID, updatedParams)
}
