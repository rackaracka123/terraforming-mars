package repository_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DemoGameService shows how a service would use the simplified repositories
type DemoGameService struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
}

func NewDemoGameService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) *DemoGameService {
	return &DemoGameService{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
	}
}

// CreateGameWithPlayer demonstrates clean service composition
func (s *DemoGameService) CreateGameWithPlayer(ctx context.Context, playerName string) (model.Game, model.Player, error) {
	// Create game
	game, err := s.gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		return model.Game{}, model.Player{}, err
	}

	// Create player
	player := model.Player{
		ID:              "player-1",
		Name:            playerName,
		TerraformRating: 20,
		Resources: model.Resources{
			Credits: 45,
		},
		Production: model.Production{
			Credits: 1,
		},
		IsConnected: true,
	}

	err = s.playerRepo.Create(ctx, game.ID, player)
	if err != nil {
		return model.Game{}, model.Player{}, err
	}

	// Add player ID to game
	err = s.gameRepo.AddPlayerID(ctx, game.ID, player.ID)
	if err != nil {
		return model.Game{}, model.Player{}, err
	}

	// Get updated game
	updatedGame, err := s.gameRepo.GetByID(ctx, game.ID)
	if err != nil {
		return model.Game{}, model.Player{}, err
	}

	return updatedGame, player, nil
}

// GetGameWithPlayers demonstrates how services compose data when needed
func (s *DemoGameService) GetGameWithPlayers(ctx context.Context, gameID string) (dto.GameDto, []dto.PlayerDto, error) {
	// Get game
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return dto.GameDto{}, nil, err
	}

	// Get players separately
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return dto.GameDto{}, nil, err
	}

	// Convert to DTOs - use basic conversion for demo test
	gameDto := dto.ToGameDtoBasic(game, dto.GetPaymentConstants())
	playerDtos := dto.ToPlayerDtoSlice(players)

	return gameDto, playerDtos, nil
}

func TestCleanArchitectureIntegration(t *testing.T) {
	// Initialize logger for testing
	err := logger.Init(nil)
	if err != nil {
		t.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Shutdown()

	// Initialize clean architecture components
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	demoService := NewDemoGameService(gameRepo, playerRepo)

	ctx := context.Background()

	t.Run("Service Layer with Clean Repositories", func(t *testing.T) {
		// Create game with player using service
		game, player, err := demoService.CreateGameWithPlayer(ctx, "Alice")
		require.NoError(t, err)

		// Verify game was created correctly
		assert.Equal(t, model.GameStatusLobby, game.Status)
		assert.Contains(t, game.PlayerIDs, "player-1")
		assert.Equal(t, "player-1", game.HostPlayerID) // First player becomes host

		// Verify player was created correctly
		assert.Equal(t, "Alice", player.Name)
		assert.Equal(t, 20, player.TerraformRating)
		assert.Equal(t, 45, player.Resources.Credits)

		// Test service composition - get game with players
		gameDto, playerDtos, err := demoService.GetGameWithPlayers(ctx, game.ID)
		require.NoError(t, err)

		// Verify DTO conversion
		assert.Equal(t, game.ID, gameDto.ID)
		assert.Equal(t, dto.GameStatus(game.Status), gameDto.Status)
		// Note: ToGameDtoBasic doesn't populate CurrentPlayer/OtherPlayers, so we verify player data separately
		assert.Len(t, playerDtos, 1)
		assert.Equal(t, "Alice", playerDtos[0].Name)

		// Test granular repository updates
		err = playerRepo.UpdateTerraformRating(ctx, game.ID, "player-1", 25)
		require.NoError(t, err)

		err = gameRepo.UpdateStatus(ctx, game.ID, model.GameStatusActive)
		require.NoError(t, err)

		// Verify updates through service
		updatedGameDto, updatedPlayerDtos, err := demoService.GetGameWithPlayers(ctx, game.ID)
		require.NoError(t, err)

		assert.Equal(t, dto.GameStatusActive, updatedGameDto.Status)
		assert.Equal(t, 25, updatedPlayerDtos[0].TerraformRating)
	})

	t.Run("Multiple Players and Complex Operations", func(t *testing.T) {
		// Create game
		game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 3})
		require.NoError(t, err)

		// Add multiple players
		players := []model.Player{
			{
				ID:              "p1",
				Name:            "Alice",
				TerraformRating: 20,
				Resources:       model.Resources{Credits: 45},
				Production:      model.Production{Credits: 1},
				IsConnected:     true,
			},
			{
				ID:              "p2",
				Name:            "Bob",
				TerraformRating: 20,
				Resources:       model.Resources{Credits: 45},
				Production:      model.Production{Credits: 1},
				IsConnected:     true,
			},
		}

		for _, player := range players {
			err = playerRepo.Create(ctx, game.ID, player)
			require.NoError(t, err)

			err = gameRepo.AddPlayerID(ctx, game.ID, player.ID)
			require.NoError(t, err)
		}

		// Test bulk operations
		gameDto, playerDtos, err := demoService.GetGameWithPlayers(ctx, game.ID)
		require.NoError(t, err)

		// Verify players exist in game through repository
		retrievedPlayers, err := playerRepo.ListByGameID(ctx, game.ID)
		require.NoError(t, err)
		assert.Len(t, retrievedPlayers, 2)
		assert.Len(t, playerDtos, 2)
		assert.Equal(t, "p1", gameDto.HostPlayerID) // First player is host

		// Test granular updates on multiple players
		for i, player := range players {
			newCredits := 50 + i*10
			err = playerRepo.UpdateResources(ctx, game.ID, player.ID, model.Resources{Credits: newCredits})
			require.NoError(t, err)
		}

		// Verify all updates
		updatedPlayers, err := playerRepo.ListByGameID(ctx, game.ID)
		require.NoError(t, err)
		assert.Len(t, updatedPlayers, 2)

		// Players should have different credit amounts
		creditAmounts := make(map[string]int)
		for _, p := range updatedPlayers {
			creditAmounts[p.ID] = p.Resources.Credits
		}
		assert.Equal(t, 50, creditAmounts["p1"])
		assert.Equal(t, 60, creditAmounts["p2"])
	})
}
