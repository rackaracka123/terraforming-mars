package lobby

import (
	"context"
	"fmt"
	"math/rand"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles all lobby operations including game creation, player joining, and game starting
type Service interface {
	// Create a new game with specified settings
	CreateGame(ctx context.Context, settings model.GameSettings) (model.Game, error)

	// Get game by ID
	GetGame(ctx context.Context, gameID string) (model.Game, error)

	// List games by status
	ListGames(ctx context.Context, status string) ([]model.Game, error)

	// Start a game (transition from status "lobby" to "active")
	StartGame(ctx context.Context, gameID string, playerID string) error

	// Add player to game (join game flow)
	JoinGame(ctx context.Context, gameID string, playerName string) (model.Game, error)
	JoinGameWithPlayerID(ctx context.Context, gameID string, playerName string, playerID string) (model.Game, error)

	// Helper methods
	AddPlayerToGame(ctx context.Context, gameID string, player model.Player) error
	GetPlayerFromGame(ctx context.Context, gameID, playerID string) (model.Player, error)
	IsGameFull(ctx context.Context, gameID string) (bool, error)
	IsHost(game *model.Game, playerID string) bool
}

// ServiceImpl implements the lobby Service interface
type ServiceImpl struct {
	gameRepo       repository.GameRepository
	playerRepo     repository.PlayerRepository
	cardRepo       repository.CardRepository
	cardService    service.CardService
	cardDeckRepo   repository.CardDeckRepository
	boardService   service.BoardService
	sessionManager session.SessionManager
}

// NewService creates a new lobby Service instance
func NewService(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	cardRepo repository.CardRepository,
	cardService service.CardService,
	cardDeckRepo repository.CardDeckRepository,
	boardService service.BoardService,
	sessionManager session.SessionManager,
) Service {
	return &ServiceImpl{
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		cardRepo:       cardRepo,
		cardService:    cardService,
		cardDeckRepo:   cardDeckRepo,
		boardService:   boardService,
		sessionManager: sessionManager,
	}
}

// CreateGame creates a new game with specified settings
func (s *ServiceImpl) CreateGame(ctx context.Context, settings model.GameSettings) (model.Game, error) {
	log := logger.WithContext()

	// Validate settings
	if err := s.validateGameSettings(settings); err != nil {
		log.Error("Invalid game settings", zap.Error(err))
		return model.Game{}, fmt.Errorf("invalid game settings: %w", err)
	}

	log.Debug("Creating game via LobbyService")

	game, err := s.gameRepo.Create(ctx, settings)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to create game: %w", err)
	}

	// Initialize the board with default Mars tiles
	board := s.boardService.GenerateDefaultBoard()
	game.Board = board

	// Update the game with the initialized board
	if err := s.gameRepo.UpdateBoard(ctx, game.ID, board); err != nil {
		log.Error("Failed to update game with board", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to initialize game board: %w", err)
	}

	// Initialize the card deck with project cards filtered by selected packs
	allProjectCards, err := s.cardRepo.GetProjectCards(ctx)
	if err != nil {
		log.Error("Failed to get project cards for deck initialization", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to get project cards: %w", err)
	}

	// Filter cards by selected packs
	selectedPacks := settings.CardPacks
	if len(selectedPacks) == 0 {
		selectedPacks = model.DefaultCardPacks()
	}

	// Create a set of selected packs for fast lookup
	packSet := make(map[string]bool)
	for _, pack := range selectedPacks {
		packSet[pack] = true
	}

	// Filter project cards by pack
	projectCards := make([]model.Card, 0)
	for _, card := range allProjectCards {
		if packSet[card.Pack] {
			projectCards = append(projectCards, card)
		}
	}

	log.Info("Filtered project cards by packs",
		zap.Strings("selected_packs", selectedPacks),
		zap.Int("total_available", len(allProjectCards)),
		zap.Int("filtered_count", len(projectCards)))

	if err := s.cardDeckRepo.InitializeDeck(ctx, game.ID, projectCards); err != nil {
		log.Error("Failed to initialize card deck", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to initialize card deck: %w", err)
	}

	log.Info("Game created via LobbyService",
		zap.String("game_id", game.ID),
		zap.Int("board_tiles", len(board.Tiles)),
		zap.Int("deck_size", len(projectCards)))
	return game, nil
}

// GetGame retrieves a game by ID
func (s *ServiceImpl) GetGame(ctx context.Context, gameID string) (model.Game, error) {
	return s.gameRepo.GetByID(ctx, gameID)
}

// ListGames lists games by status
func (s *ServiceImpl) ListGames(ctx context.Context, status string) ([]model.Game, error) {
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
	if game.Status != model.GameStatusLobby {
		log.Warn("Attempted to start game not in lobby state", zap.String("current_status", string(game.Status)))
		return fmt.Errorf("game is not in lobby state")
	}
	if len(players) < 1 {
		log.Warn("Attempted to start game with no players")
		return fmt.Errorf("cannot start game with no players")
	}

	// Transition game status to active using granular updates
	if err := s.gameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive); err != nil {
		log.Error("Failed to update game status to active", zap.Error(err))
		return fmt.Errorf("failed to update game status: %w", err)
	}

	if err := s.gameRepo.UpdatePhase(ctx, gameID, model.GamePhaseStartingCardSelection); err != nil {
		log.Error("Failed to update game phase", zap.Error(err))
		return fmt.Errorf("failed to update game phase: %w", err)
	}

	// Set the first player's turn (typically the host)
	if len(players) > 0 {
		firstPlayerID := players[0].ID
		if err := s.gameRepo.SetCurrentTurn(ctx, gameID, &firstPlayerID); err != nil {
			log.Error("Failed to set initial current turn", zap.Error(err))
			return fmt.Errorf("failed to set current turn: %w", err)
		}
		log.Info("Set initial current turn", zap.String("first_player_id", firstPlayerID))
	}

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

	// Broadcast game state to all players after starting
	err = s.sessionManager.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast game started", zap.Error(err))
		// Don't fail the start operation, just log the error
	}

	return nil
}

// JoinGame adds a player to a game using both GameState and Player repositories
func (s *ServiceImpl) JoinGame(ctx context.Context, gameID string, playerName string) (model.Game, error) {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Player joining game via LobbyService", zap.String("player_name", playerName))

	// Get the current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for join", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to get game: %w", err)
	}

	// Check if game is joinable
	if game.Status == model.GameStatusCompleted {
		log.Warn("Attempted to join completed game", zap.String("player_name", playerName))
		return model.Game{}, fmt.Errorf("cannot join completed game")
	}

	// Check if game is full
	isFull, err := s.IsGameFull(ctx, gameID)
	if err != nil {
		log.Error("Failed to check if game is full", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to check game capacity: %w", err)
	}
	if isFull {
		log.Warn("Attempted to join full game", zap.String("player_name", playerName))
		return model.Game{}, fmt.Errorf("game is full")
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
		PlayedCards:      make([]string, 0),
		Passed:           false, // Player starts active, not passed
		AvailableActions: 2,     // Standard actions per turn in action phase
		VictoryPoints:    0,     // Starting victory points
		IsConnected:      true,
	}

	// Add player using clean architecture method
	if err := s.AddPlayerToGame(ctx, gameID, player); err != nil {
		log.Error("Failed to add player to game", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to add player: %w", err)
	}

	// Host setting is handled automatically by gameRepo.AddPlayerID
	// Get updated game state to return
	updatedGame, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to get updated game: %w", err)
	}

	// If game is in starting_card_selection phase, distribute starting cards to the new player
	if updatedGame.Status == model.GameStatusActive && updatedGame.CurrentPhase == model.GamePhaseStartingCardSelection {
		log.Debug("Game is in starting card selection phase, distributing cards to new player", zap.String("player_id", playerID))

		// Create a slice with just the new player for card distribution
		newPlayerSlice := []model.Player{player}

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
func (s *ServiceImpl) JoinGameWithPlayerID(ctx context.Context, gameID string, playerName string, playerID string) (model.Game, error) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player joining game via LobbyService with pre-specified ID", zap.String("player_name", playerName))

	// Get the current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for join", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to get game: %w", err)
	}

	// Check if game is joinable
	if game.Status == model.GameStatusCompleted {
		log.Warn("Attempted to join completed game", zap.String("player_name", playerName))
		return model.Game{}, fmt.Errorf("cannot join completed game")
	}

	// Check if game is full
	isFull, err := s.IsGameFull(ctx, gameID)
	if err != nil {
		log.Error("Failed to check if game is full", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to check game capacity: %w", err)
	}
	if isFull {
		log.Warn("Attempted to join full game", zap.String("player_name", playerName))
		return model.Game{}, fmt.Errorf("game is full")
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
		PlayedCards:      make([]string, 0),
		Passed:           false, // Player starts active, not passed
		AvailableActions: 2,     // Standard actions per turn in action phase
		VictoryPoints:    0,     // Starting victory points
		IsConnected:      true,
	}

	// Add player using clean architecture method (same as in JoinGame)
	if err := s.AddPlayerToGame(ctx, gameID, player); err != nil {
		log.Error("Failed to add player to game", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to add player: %w", err)
	}

	// Host setting is handled automatically by gameRepo.AddPlayerID

	// Get updated game state
	updatedGame, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to get updated game: %w", err)
	}

	// If game is in starting_card_selection phase, distribute starting cards to the new player
	if updatedGame.Status == model.GameStatusActive && updatedGame.CurrentPhase == model.GamePhaseStartingCardSelection {
		log.Debug("Game is in starting card selection phase, distributing cards to new player", zap.String("player_id", playerID))

		// Create a slice with just the new player for card distribution
		newPlayerSlice := []model.Player{player}

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
func (s *ServiceImpl) AddPlayerToGame(ctx context.Context, gameID string, player model.Player) error {
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
func (s *ServiceImpl) GetPlayerFromGame(ctx context.Context, gameID, playerID string) (model.Player, error) {
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
func (s *ServiceImpl) IsHost(game *model.Game, playerID string) bool {
	return game.HostPlayerID == playerID
}

// distributeStartingCards gives each player 10 random cards and 2 corporations to choose from for starting selection
func (s *ServiceImpl) distributeStartingCards(ctx context.Context, gameID string, players []model.Player) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Distributing starting cards and corporations to players")

	// Get game to access settings
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Get the starting card pool from card service
	startingCards, err := s.cardService.GetStartingCards(ctx)
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
		selectedPacks = model.DefaultCardPacks()
	}

	// Create a set of selected packs for fast lookup
	packSet := make(map[string]bool)
	for _, pack := range selectedPacks {
		packSet[pack] = true
	}

	// Filter corporations by pack
	allCorporations := make([]model.Card, 0)
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
	cardsWithRequirements := make([]model.Card, 0)
	cardsWithoutRequirements := make([]model.Card, 0)

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
		playerStartingCards := make([]model.Card, 0, 10)

		// Always include at least one card with requirements (if available)
		if len(cardsWithRequirements) > 0 {
			// Shuffle cards with requirements and pick one
			requirementCards := make([]model.Card, len(cardsWithRequirements))
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
			remainingPool := make([]model.Card, 0, len(startingCards)-len(playerStartingCards))
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
		corporationPool := make([]model.Card, len(allCorporations))
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
		startingCardsPhase := model.SelectStartingCardsPhase{
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
func (s *ServiceImpl) validateGameSettings(settings model.GameSettings) error {
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
