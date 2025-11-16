package lobby

import (
	"context"
	"fmt"
	"math/rand"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/features/turn"
	gameModel "terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	playerPkg "terraforming-mars-backend/internal/player"
	sessionPkg "terraforming-mars-backend/internal/session"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles all lobby operations including game creation, player joining, and game starting.
//
// Scope: Pre-game phase operations ONLY
//   - Game creation with settings validation
//   - Player joining and capacity management
//   - Host management and permissions
//   - Game starting (lobby → active transition)
//   - Starting card distribution (as part of game start)
//
// Boundary: Once game.Status = "active", all operations move to GameService and other services
// This service should NOT handle active gameplay operations like:
//   - Turn management
//   - Card playing
//   - Resource/production updates
//   - Tile placement
//   - Global parameter changes
type Service interface {
	// Create a new game with specified settings
	CreateGame(ctx context.Context, settings gameModel.GameSettings) (gameModel.Game, error)

	// Get game by ID
	GetGame(ctx context.Context, gameID string) (gameModel.Game, error)

	// List games by status
	ListGames(ctx context.Context, status string) ([]gameModel.Game, error)

	// Start a game (transition from status "lobby" to "active")
	StartGame(ctx context.Context, gameID string, playerID string) error

	// Add player to game (join game flow)
	JoinGame(ctx context.Context, gameID string, playerName string) (gameModel.Game, error)
	JoinGameWithPlayerID(ctx context.Context, gameID string, playerName string, playerID string) (gameModel.Game, error)

	// Helper methods
	AddPlayerToGame(ctx context.Context, gameID string, player playerPkg.Player) error
	GetPlayerFromGame(ctx context.Context, gameID, playerID string) (playerPkg.Player, error)
	IsGameFull(ctx context.Context, gameID string) (bool, error)
	IsHost(game *gameModel.Game, playerID string) bool
}

// ServiceImpl implements the lobby Service interface
type ServiceImpl struct {
	gameRepo       gameModel.Repository
	playerRepo     playerPkg.Repository
	cardRepo       card.CardRepository
	cardDeckRepo   card.CardDeckRepository
	sessionManager session.SessionManager
	sessionRepo    sessionPkg.Repository
	eventBus       *events.EventBusImpl
}

// NewService creates a new lobby Service instance
func NewService(
	gameRepo gameModel.Repository,
	playerRepo playerPkg.Repository,
	cardRepo card.CardRepository,
	cardDeckRepo card.CardDeckRepository,
	sessionManager session.SessionManager,
	sessionRepo sessionPkg.Repository,
	eventBus *events.EventBusImpl,
) Service {
	return &ServiceImpl{
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		cardRepo:       cardRepo,
		cardDeckRepo:   cardDeckRepo,
		sessionManager: sessionManager,
		sessionRepo:    sessionRepo,
		eventBus:       eventBus,
	}
}

// CreateGame creates a new game with specified settings
func (s *ServiceImpl) CreateGame(ctx context.Context, settings gameModel.GameSettings) (gameModel.Game, error) {
	log := logger.WithContext()

	// Validate settings
	if err := s.validateGameSettings(settings); err != nil {
		log.Error("Invalid game settings", zap.Error(err))
		return gameModel.Game{}, fmt.Errorf("invalid game settings: %w", err)
	}

	log.Debug("Creating game via LobbyService")

	newGame, err := s.gameRepo.Create(ctx, settings)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		return gameModel.Game{}, fmt.Errorf("failed to create game: %w", err)
	}

	// Board will be initialized when session is created (in StartGame)
	// No need to store board in game metadata

	// Initialize the card deck with project cards filtered by selected packs
	allProjectCards, err := s.cardRepo.GetProjectCards(ctx)
	if err != nil {
		log.Error("Failed to get project cards for deck initialization", zap.Error(err))
		return gameModel.Game{}, fmt.Errorf("failed to get project cards: %w", err)
	}

	// Filter cards by selected packs
	selectedPacks := settings.CardPacks
	if len(selectedPacks) == 0 {
		selectedPacks = gameModel.DefaultCardPacks()
	}

	// Create a set of selected packs for fast lookup
	packSet := make(map[string]bool)
	for _, pack := range selectedPacks {
		packSet[pack] = true
	}

	// Filter project cards by pack
	projectCards := make([]card.Card, 0)
	for _, c := range allProjectCards {
		if packSet[c.Pack] {
			projectCards = append(projectCards, c)
		}
	}

	log.Info("Filtered project cards by packs",
		zap.Strings("selected_packs", selectedPacks),
		zap.Int("total_available", len(allProjectCards)),
		zap.Int("filtered_count", len(projectCards)))

	if err := s.cardDeckRepo.InitializeDeck(ctx, newGame.ID, projectCards); err != nil {
		log.Error("Failed to initialize card deck", zap.Error(err))
		return gameModel.Game{}, fmt.Errorf("failed to initialize card deck: %w", err)
	}

	log.Info("Game created via LobbyService",
		zap.String("game_id", newGame.ID),
		zap.Int("deck_size", len(projectCards)))
	return newGame, nil
}

// GetGame retrieves a game by ID
func (s *ServiceImpl) GetGame(ctx context.Context, gameID string) (gameModel.Game, error) {
	return s.gameRepo.GetByID(ctx, gameID)
}

// ListGames lists games by status
func (s *ServiceImpl) ListGames(ctx context.Context, status string) ([]gameModel.Game, error) {
	return s.gameRepo.List(ctx, status)
}

// StartGame transitions a game from lobby to active status
func (s *ServiceImpl) StartGame(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Starting game via LobbyService")

	// Get current game state to validate
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for start", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Get players separately
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players for start", zap.Error(err))
		return fmt.Errorf("failed to get players: %w", err)
	}

	// Ensure player is host
	if game.HostPlayerID != playerID {
		log.Warn("Non-host player attempted to start game", zap.String("player_id", playerID))
		return fmt.Errorf("only the host can start the game")
	}

	// Validate game can be started
	if game.Status != gameModel.GameStatusLobby {
		log.Warn("Attempted to start game not in lobby state", zap.String("current_status", string(game.Status)))
		return fmt.Errorf("game is not in lobby state")
	}
	if len(players) < 1 {
		log.Warn("Attempted to start game with no players")
		return fmt.Errorf("cannot start game with no players")
	}

	// Transition game status to active using granular updates
	if err := s.gameRepo.UpdateStatus(ctx, gameID, gameModel.GameStatusActive); err != nil {
		log.Error("Failed to update game status to active", zap.Error(err))
		return fmt.Errorf("failed to update game status: %w", err)
	}

	if err := s.gameRepo.UpdatePhase(ctx, gameID, gameModel.GamePhaseStartingCardSelection); err != nil {
		log.Error("Failed to update game phase", zap.Error(err))
		return fmt.Errorf("failed to update game phase: %w", err)
	}

	// Feature repositories will be created below when setting up the session

	// Distribute starting cards to all players
	if err := s.distributeStartingCards(ctx, gameID, players); err != nil {
		log.Error("Failed to distribute starting cards", zap.Error(err))
		return fmt.Errorf("failed to distribute starting cards: %w", err)
	}

	// If we are starting a game with 1 player, give it unlimited actions (-1)
	if len(players) == 1 {
		soloPlayerID := players[0].ID
		if err := s.playerRepo.UpdateAvailableActions(ctx, gameID, soloPlayerID, -1); err != nil {
			log.Error("Failed to set unlimited actions for solo player", zap.Error(err))
		}
		log.Info("🏃 Solo player detected at game start - granting unlimited actions",
			zap.String("player_id", soloPlayerID),
			zap.Int("total_players", len(players)))
	}

	log.Info("Game started", zap.String("game_id", gameID))

	// Create runtime session for the active game
	// Get updated game
	game, err = s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for session creation", zap.Error(err))
		return fmt.Errorf("failed to get game for session: %w", err)
	}

	// Get updated players (now with starting cards distributed)
	players, err = s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players for session creation", zap.Error(err))
		return fmt.Errorf("failed to get players for session: %w", err)
	}

	// Convert player slice to map
	playersMap := make(map[string]*playerPkg.Player)
	for i := range players {
		playersMap[players[i].ID] = &players[i]
	}

	// Create feature repositories for this game session
	// These are created here (in StartGame) rather than in GameRepository.Create
	// because they're runtime concerns, not persistent game metadata

	// Parameters repository (temperature, oxygen, oceans)
	// Apply defaults if not specified (Settings uses pointers)
	temperature := gameModel.DefaultTemperature
	if game.Settings.Temperature != nil {
		temperature = *game.Settings.Temperature
	}
	oxygen := gameModel.DefaultOxygen
	if game.Settings.Oxygen != nil {
		oxygen = *game.Settings.Oxygen
	}
	oceans := gameModel.DefaultOceans
	if game.Settings.Oceans != nil {
		oceans = *game.Settings.Oceans
	}

	parametersRepo, err := parameters.NewRepository(gameID, parameters.GlobalParameters{
		Temperature: temperature,
		Oxygen:      oxygen,
		Oceans:      oceans,
	}, s.eventBus)
	if err != nil {
		log.Error("Failed to create parameters repository", zap.Error(err))
		return fmt.Errorf("failed to create parameters repository: %w", err)
	}
	parametersService := parameters.NewService(parametersRepo)

	// Board repository and service (created per-game when game starts)
	// Generate default board using a temporary service, then create repository
	tempBoardService := tiles.NewBoardService(nil)
	defaultBoard := tempBoardService.GenerateDefaultBoard()
	boardRepo := tiles.NewBoardRepository(gameID, defaultBoard, s.eventBus)
	boardService := tiles.NewBoardService(boardRepo)

	// Turn order repository and service
	playerIDs := make([]string, len(players))
	for i, p := range players {
		playerIDs[i] = p.ID
	}
	var firstPlayerID *string
	if len(playerIDs) > 0 {
		firstPlayerID = &playerIDs[0]
	}
	turnOrderRepo := turn.NewTurnOrderRepository(playerIDs, firstPlayerID)
	turnOrderService := turn.NewTurnOrderService(turnOrderRepo)

	log.Info("Created feature repositories for game session",
		zap.Int("temperature", temperature),
		zap.Int("oxygen", oxygen),
		zap.Int("oceans", oceans),
		zap.Int("player_count", len(playerIDs)))

	// Create session with feature services
	gameSession := sessionPkg.NewSession(
		gameID,
		&game,
		playersMap,
		parametersService,
		boardService,
		nil, // CardService - will be injected later when needed
		turnOrderService,
		nil, // GreeneryRuleSubscriber removed - greenery logic handled directly in tile placement
		game.HostPlayerID,
	)

	// Add session to repository
	if err := s.sessionRepo.Add(ctx, gameSession); err != nil {
		log.Error("Failed to add session to repository", zap.Error(err))
		return fmt.Errorf("failed to create session: %w", err)
	}

	log.Info("✅ Runtime session created for active game",
		zap.String("game_id", gameID),
		zap.Int("player_count", len(playersMap)))

	return nil
}

// JoinGame adds a player to a game using both GameState and Player repositories
func (s *ServiceImpl) JoinGame(ctx context.Context, gameID string, playerName string) (gameModel.Game, error) {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Player joining game via LobbyService", zap.String("player_name", playerName))

	// Get the current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for join", zap.Error(err))
		return gameModel.Game{}, fmt.Errorf("failed to get game: %w", err)
	}

	// Check if game is joinable
	if game.Status == gameModel.GameStatusCompleted {
		log.Warn("Attempted to join completed game", zap.String("player_name", playerName))
		return gameModel.Game{}, fmt.Errorf("cannot join completed game")
	}

	// Check if game is full
	isFull, err := s.IsGameFull(ctx, gameID)
	if err != nil {
		log.Error("Failed to check if game is full", zap.Error(err))
		return gameModel.Game{}, fmt.Errorf("failed to check game capacity: %w", err)
	}
	if isFull {
		log.Warn("Attempted to join full game", zap.String("player_name", playerName))
		return gameModel.Game{}, fmt.Errorf("game is full")
	}

	// Check if player with this name already exists to prevent duplicates
	existingPlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players for duplicate check", zap.Error(err))
		// Continue with player creation if we can't check for duplicates
	} else {
		for _, player := range existingPlayers {
			if player.Name == playerName {
				log.Debug("Player with this name already exists, returning existing player",
					zap.String("existing_player_id", player.ID),
					zap.String("player_name", playerName))
				// Return existing game state since player already exists
				return game, nil
			}
		}
	}

	// Create new player
	playerID := uuid.New().String()
	player := playerPkg.Player{
		ID:                 playerID,
		Name:               playerName,
		TerraformRating:    20,                        // Starting terraform rating
		Cards:              make([]playerPkg.Card, 0), // Empty hand (Card instances)
		PlayedCards:        make([]playerPkg.Card, 0), // No cards played yet (Card instances)
		VictoryPoints:      0,                         // Starting victory points
		IsConnected:        true,
		Effects:            make([]playerPkg.PlayerEffect, 0),      // No effects initially
		Actions:            make([]playerPkg.PlayerAction, 0),      // No actions initially
		ResourceStorage:    make(map[string]int),                   // Empty resource storage
		PaymentSubstitutes: make([]playerPkg.PaymentSubstitute, 0), // No substitutes initially
		// Services will be injected after creation by repository
	}

	// Add player using clean architecture method
	if err := s.AddPlayerToGame(ctx, gameID, player); err != nil {
		log.Error("Failed to add player to game", zap.Error(err))
		return gameModel.Game{}, fmt.Errorf("failed to add player: %w", err)
	}

	// Host setting is handled automatically by gameRepo.AddPlayerID
	// Get updated game state to return
	updatedGame, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game", zap.Error(err))
		return gameModel.Game{}, fmt.Errorf("failed to get updated game: %w", err)
	}

	// If game is in starting_card_selection phase, distribute starting cards to the new player
	if updatedGame.Status == gameModel.GameStatusActive && updatedGame.CurrentPhase == gameModel.GamePhaseStartingCardSelection {
		log.Debug("Game is in starting card selection phase, distributing cards to new player", zap.String("player_id", playerID))

		// Create a slice with just the new player for card distribution
		newPlayerSlice := []playerPkg.Player{player}

		if err := s.distributeStartingCards(ctx, gameID, newPlayerSlice); err != nil {
			log.Error("Failed to distribute starting cards to new player", zap.Error(err), zap.String("player_id", playerID))
			// Don't return error here - player joined successfully, just missing starting cards
			// We could handle this gracefully by allowing them to get cards later
		} else {
			log.Info("🃏 Distributed starting cards to late-joining player", zap.String("player_id", playerID))
		}
	}

	log.Debug("Player joined game", zap.String("player_id", playerID))

	// Broadcast updated game state to all players
	err = s.sessionManager.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast player joined", zap.Error(err))
		// Don't fail the join operation, just log the error
	}

	return updatedGame, nil
}

// JoinGameWithPlayerID allows a player to join with a pre-specified player ID
// This is used by WebSocket connections to ensure consistent player ID before broadcasting
func (s *ServiceImpl) JoinGameWithPlayerID(ctx context.Context, gameID string, playerName string, playerID string) (gameModel.Game, error) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player joining game via LobbyService with pre-specified ID", zap.String("player_name", playerName))

	// Get the current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for join", zap.Error(err))
		return gameModel.Game{}, fmt.Errorf("failed to get game: %w", err)
	}

	// Check if game is joinable
	if game.Status == gameModel.GameStatusCompleted {
		log.Warn("Attempted to join completed game", zap.String("player_name", playerName))
		return gameModel.Game{}, fmt.Errorf("cannot join completed game")
	}

	// Check if game is full
	isFull, err := s.IsGameFull(ctx, gameID)
	if err != nil {
		log.Error("Failed to check if game is full", zap.Error(err))
		return gameModel.Game{}, fmt.Errorf("failed to check game capacity: %w", err)
	}
	if isFull {
		log.Warn("Attempted to join full game", zap.String("player_name", playerName))
		return gameModel.Game{}, fmt.Errorf("game is full")
	}

	// Check if player with this name already exists to prevent duplicates
	existingPlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players for duplicate check", zap.Error(err))
		// Continue with player creation if we can't check for duplicates
	} else {
		for _, player := range existingPlayers {
			if player.Name == playerName {
				log.Debug("Player with this name already exists, returning existing player",
					zap.String("existing_player_id", player.ID),
					zap.String("player_name", playerName))
				// Return existing game state since player already exists
				return game, nil
			}
		}
	}

	// Create new player with the provided ID
	player := playerPkg.Player{
		ID:                 playerID,
		Name:               playerName,
		TerraformRating:    20,                        // Starting terraform rating
		Cards:              make([]playerPkg.Card, 0), // Empty hand (Card instances)
		PlayedCards:        make([]playerPkg.Card, 0), // No cards played yet (Card instances)
		VictoryPoints:      0,                         // Starting victory points
		IsConnected:        true,
		Effects:            make([]playerPkg.PlayerEffect, 0),      // No effects initially
		Actions:            make([]playerPkg.PlayerAction, 0),      // No actions initially
		ResourceStorage:    make(map[string]int),                   // Empty resource storage
		PaymentSubstitutes: make([]playerPkg.PaymentSubstitute, 0), // No substitutes initially
		// Services will be injected after creation by repository
	}

	// Add player using clean architecture method (same as in JoinGame)
	if err := s.AddPlayerToGame(ctx, gameID, player); err != nil {
		log.Error("Failed to add player to game", zap.Error(err))
		return gameModel.Game{}, fmt.Errorf("failed to add player: %w", err)
	}

	// Host setting is handled automatically by gameRepo.AddPlayerID

	// Get updated game state
	updatedGame, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game", zap.Error(err))
		return gameModel.Game{}, fmt.Errorf("failed to get updated game: %w", err)
	}

	// If game is in starting_card_selection phase, distribute starting cards to the new player
	if updatedGame.Status == gameModel.GameStatusActive && updatedGame.CurrentPhase == gameModel.GamePhaseStartingCardSelection {
		log.Debug("Game is in starting card selection phase, distributing cards to new player", zap.String("player_id", playerID))

		// Create a slice with just the new player for card distribution
		newPlayerSlice := []playerPkg.Player{player}

		if err := s.distributeStartingCards(ctx, gameID, newPlayerSlice); err != nil {
			log.Error("Failed to distribute starting cards to new player", zap.Error(err), zap.String("player_id", playerID))
			// Don't return error here - player joined successfully, just missing starting cards
			// We could handle this gracefully by allowing them to get cards later
		} else {
			log.Info("🃏 Distributed starting cards to late-joining player", zap.String("player_id", playerID))
		}
	}

	log.Debug("Player joined game", zap.String("player_id", playerID))

	// Broadcast updated game state to all players
	err = s.sessionManager.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast player joined", zap.Error(err))
		// Don't fail the join operation, just log the error
	}

	return updatedGame, nil
}

// AddPlayerToGame adds a player to the game (clean architecture pattern)
func (s *ServiceImpl) AddPlayerToGame(ctx context.Context, gameID string, player playerPkg.Player) error {
	// Get current game to check max players
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return err
	}

	// Check if game is full
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return err
	}

	if len(players) >= game.Settings.MaxPlayers {
		return fmt.Errorf("game is full")
	}

	// Add player to player repository
	if err := s.playerRepo.Create(ctx, gameID, player); err != nil {
		return err
	}

	// Add player ID to game
	return s.gameRepo.AddPlayerID(ctx, gameID, player.ID)
}

// GetPlayerFromGame returns a player by ID (clean architecture pattern)
func (s *ServiceImpl) GetPlayerFromGame(ctx context.Context, gameID, playerID string) (playerPkg.Player, error) {
	return s.playerRepo.GetByID(ctx, gameID, playerID)
}

// IsGameFull returns true if the game has reached maximum players (clean architecture pattern)
func (s *ServiceImpl) IsGameFull(ctx context.Context, gameID string) (bool, error) {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return false, err
	}

	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return false, err
	}

	return len(players) >= game.Settings.MaxPlayers, nil
}

// IsHost returns true if the given player ID is the host of the game (business logic from Game model)
func (s *ServiceImpl) IsHost(game *gameModel.Game, playerID string) bool {
	return game.HostPlayerID == playerID
}

// distributeStartingCards gives each player 10 random cards and 2 corporations to choose from for starting selection
func (s *ServiceImpl) distributeStartingCards(ctx context.Context, gameID string, players []playerPkg.Player) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Distributing starting cards and corporations to players")

	// Get game to access settings
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Get the starting card pool from card repository
	startingCards, err := s.cardRepo.GetStartingCardPool(ctx)
	if err != nil {
		return fmt.Errorf("failed to get starting card pool: %w", err)
	}

	if len(startingCards) < 10 {
		return fmt.Errorf("insufficient cards in pool: need at least 10, got %d", len(startingCards))
	}

	// Get all corporation cards and filter by selected packs
	allCorporationCards, err := s.cardRepo.GetCorporationCards(ctx)
	if err != nil {
		return fmt.Errorf("failed to get corporation cards: %w", err)
	}

	// Filter corporations by selected packs
	selectedPacks := game.Settings.CardPacks
	if len(selectedPacks) == 0 {
		selectedPacks = gameModel.DefaultCardPacks()
	}

	// Create a set of selected packs for fast lookup
	packSet := make(map[string]bool)
	for _, pack := range selectedPacks {
		packSet[pack] = true
	}

	// Filter corporations by pack
	allCorporations := make([]card.Card, 0)
	for _, corp := range allCorporationCards {
		if packSet[corp.Pack] {
			allCorporations = append(allCorporations, corp)
		}
	}

	log.Info("Filtered corporations by packs",
		zap.Strings("selected_packs", selectedPacks),
		zap.Int("total_available", len(allCorporationCards)),
		zap.Int("filtered_count", len(allCorporations)))

	if len(allCorporations) < 2 {
		return fmt.Errorf("insufficient corporations in selected packs: need at least 2, got %d", len(allCorporations))
	}

	// Separate cards into groups: cards with requirements and without requirements
	cardsWithRequirements := make([]card.Card, 0)
	cardsWithoutRequirements := make([]card.Card, 0)

	for _, card := range startingCards {
		hasRequirements := len(card.Requirements) > 0

		if hasRequirements {
			cardsWithRequirements = append(cardsWithRequirements, card)
		} else {
			cardsWithoutRequirements = append(cardsWithoutRequirements, card)
		}
	}

	log.Debug("Card pools separated",
		zap.Int("cards_with_requirements", len(cardsWithRequirements)),
		zap.Int("cards_without_requirements", len(cardsWithoutRequirements)))

	// For each player, select 10 cards: at least 1 with requirements, rest random
	for _, player := range players {
		playerStartingCards := make([]card.Card, 0, 10)

		// Always include at least one card with requirements (if available)
		if len(cardsWithRequirements) > 0 {
			// Shuffle cards with requirements and pick one
			requirementCards := make([]card.Card, len(cardsWithRequirements))
			copy(requirementCards, cardsWithRequirements)
			for i := len(requirementCards) - 1; i > 0; i-- {
				j := rand.Intn(i + 1)
				requirementCards[i], requirementCards[j] = requirementCards[j], requirementCards[i]
			}
			playerStartingCards = append(playerStartingCards, requirementCards[0])
		}

		// Fill remaining slots with random cards from the entire pool
		remainingSlots := 10 - len(playerStartingCards)
		if remainingSlots > 0 {
			// Create a pool of remaining cards (excluding already selected)
			remainingPool := make([]card.Card, 0, len(startingCards)-len(playerStartingCards))
			selectedIDs := make(map[string]bool)
			for _, card := range playerStartingCards {
				selectedIDs[card.ID] = true
			}

			for _, card := range startingCards {
				if !selectedIDs[card.ID] {
					remainingPool = append(remainingPool, card)
				}
			}

			// Shuffle remaining pool and take what we need
			for i := len(remainingPool) - 1; i > 0; i-- {
				j := rand.Intn(i + 1)
				remainingPool[i], remainingPool[j] = remainingPool[j], remainingPool[i]
			}

			// Add remaining cards up to 10 total
			cardsToAdd := remainingSlots
			if cardsToAdd > len(remainingPool) {
				cardsToAdd = len(remainingPool)
			}
			playerStartingCards = append(playerStartingCards, remainingPool[:cardsToAdd]...)
		}

		// Convert cards to card IDs
		playerStartingCardIDs := make([]string, len(playerStartingCards))
		for i, card := range playerStartingCards {
			playerStartingCardIDs[i] = card.ID
		}

		// Select 2 random corporations for this player
		corporationPool := make([]card.Card, len(allCorporations))
		copy(corporationPool, allCorporations)

		// Shuffle corporations
		for i := len(corporationPool) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			corporationPool[i], corporationPool[j] = corporationPool[j], corporationPool[i]
		}

		// Take first 2 corporations
		playerCorporationIDs := make([]string, 2)
		playerCorporationIDs[0] = corporationPool[0].ID
		playerCorporationIDs[1] = corporationPool[1].ID

		// Update the player with starting card IDs and corporation options
		startingCardsPhase := playerPkg.SelectStartingCardsPhase{
			AvailableCards:        playerStartingCardIDs,
			AvailableCorporations: playerCorporationIDs,
		}

		if err := s.playerRepo.UpdateSelectStartingCardsPhase(ctx, gameID, player.ID, &startingCardsPhase); err != nil {
			log.Error("Failed to set starting selection for player",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return fmt.Errorf("failed to set starting selection for player %s: %w", player.ID, err)
		}

		log.Info("Distributed starting cards and corporations to player",
			zap.String("player_id", player.ID),
			zap.Int("card_count", len(playerStartingCards)),
			zap.Strings("corporations", playerCorporationIDs))
	}

	log.Info("Successfully distributed starting cards to all players",
		zap.Int("player_count", len(players)))
	return nil
}

// validateGameSettings validates game creation settings
func (s *ServiceImpl) validateGameSettings(settings gameModel.GameSettings) error {
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
