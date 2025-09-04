package service_test

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"testing"
)

// Helper functions for creating pointers
func stringPtr(s string) *string { return &s }
func intPtr(i int) *int         { return &i }

func TestNewGameService(t *testing.T) {
	gameRepo := repository.NewGameRepository()
	eventBus := events.NewInMemoryEventBus()
	gameService := service.NewGameService(gameRepo, eventBus, nil)

	if gameService == nil {
		t.Fatal("Expected service to be non-nil")
	}
}

func TestGameService_CreateGame(t *testing.T) {
	gameRepo := repository.NewGameRepository()
	gameService := service.NewGameService(gameRepo, events.NewInMemoryEventBus(), nil)

	tests := []struct {
		name     string
		settings model.GameSettings
		wantErr  bool
	}{
		{
			name: "valid game settings",
			settings: model.GameSettings{
				MaxPlayers: 4,
			},
			wantErr: false,
		},
		{
			name: "max players too high",
			settings: model.GameSettings{
				MaxPlayers: 6,
			},
			wantErr: true,
		},
		{
			name: "max players too low",
			settings: model.GameSettings{
				MaxPlayers: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game, err := gameService.CreateGame(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateGame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if game == nil {
					t.Error("Expected game to be non-nil")
				}
				if game.Settings.MaxPlayers != tt.settings.MaxPlayers {
					t.Errorf("Expected MaxPlayers to be %d, got %d", tt.settings.MaxPlayers, game.Settings.MaxPlayers)
				}
			}
		})
	}
}

func TestGameService_GetGame(t *testing.T) {
	gameRepo := repository.NewGameRepository()
	gameService := service.NewGameService(gameRepo, events.NewInMemoryEventBus(), nil)

	// Create a game first
	settings := model.GameSettings{MaxPlayers: 4}
	game, err := gameService.CreateGame(settings)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	tests := []struct {
		name    string
		gameID  string
		wantErr bool
	}{
		{
			name:    "valid game ID",
			gameID:  game.ID,
			wantErr: false,
		},
		{
			name:    "empty game ID",
			gameID:  "",
			wantErr: true,
		},
		{
			name:    "non-existent game ID",
			gameID:  "non-existent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGame, err := gameService.GetGame(tt.gameID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotGame == nil {
					t.Error("Expected game to be non-nil")
				}
				if gotGame.ID != tt.gameID {
					t.Errorf("Expected game ID to be %s, got %s", tt.gameID, gotGame.ID)
				}
			}
		})
	}
}

func TestGameService_JoinGame(t *testing.T) {
	gameRepo := repository.NewGameRepository()
	gameService := service.NewGameService(gameRepo, events.NewInMemoryEventBus(), nil)

	// Create a game first
	settings := model.GameSettings{MaxPlayers: 2}
	game, err := gameService.CreateGame(settings)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	tests := []struct {
		name       string
		gameID     string
		playerName string
		wantErr    bool
	}{
		{
			name:       "valid join",
			gameID:     game.ID,
			playerName: "Player1",
			wantErr:    false,
		},
		{
			name:       "another valid join",
			gameID:     game.ID,
			playerName: "Player2",
			wantErr:    false,
		},
		{
			name:       "game full",
			gameID:     game.ID,
			playerName: "Player3",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGame, err := gameService.JoinGame(tt.gameID, tt.playerName)
			if (err != nil) != tt.wantErr {
				t.Errorf("JoinGame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotGame == nil {
					t.Error("Expected game to be non-nil")
				}

				// Check if player was added
				playerFound := false
				for _, player := range gotGame.Players {
					if player.Name == tt.playerName {
						playerFound = true
						// Verify player initialization
						if player.Resources.Credits != 0 {
							t.Errorf("Expected initial credits to be 0, got %d", player.Resources.Credits)
						}
						if player.Production.Credits != 1 {
							t.Errorf("Expected initial credit production to be 1, got %d", player.Production.Credits)
						}
						if player.TerraformRating != 20 {
							t.Errorf("Expected initial terraform rating to be 20, got %d", player.TerraformRating)
						}
						if !player.IsActive {
							t.Error("Expected player to be active")
						}
						if len(player.PlayedCards) != 0 {
							t.Errorf("Expected empty played cards, got %v", player.PlayedCards)
						}
						break
					}
				}
				if !playerFound {
					t.Errorf("Player %s not found in game", tt.playerName)
				}
			}
		})
	}
}

func TestGameService_ListGames(t *testing.T) {
	gameRepo := repository.NewGameRepository()
	gameService := service.NewGameService(gameRepo, events.NewInMemoryEventBus(), nil)

	// Create test games
	settings1 := model.GameSettings{MaxPlayers: 4}
	game1, err := gameService.CreateGame(settings1)
	if err != nil {
		t.Fatalf("Failed to create game1: %v", err)
	}

	settings2 := model.GameSettings{MaxPlayers: 3}
	game2, err := gameService.CreateGame(settings2)
	if err != nil {
		t.Fatalf("Failed to create game2: %v", err)
	}

	tests := []struct {
		name       string
		status     string
		expectGame bool
		gameID     string
	}{
		{
			name:       "list all games",
			status:     "",
			expectGame: true,
			gameID:     game1.ID,
		},
		{
			name:       "list lobby games",
			status:     string(model.GameStatusLobby),
			expectGame: true,
			gameID:     game2.ID,
		},
		{
			name:       "list completed games",
			status:     string(model.GameStatusCompleted),
			expectGame: false,
			gameID:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			games, err := gameService.ListGames(tt.status)
			if err != nil {
				t.Errorf("ListGames() error = %v", err)
				return
			}

			if tt.expectGame {
				if len(games) == 0 {
					t.Error("Expected at least one game")
					return
				}
				if tt.gameID != "" {
					found := false
					for _, g := range games {
						if g.ID == tt.gameID {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected to find game %s", tt.gameID)
					}
				}
			} else {
				if len(games) != 0 {
					t.Errorf("Expected no games, got %d", len(games))
				}
			}
		})
	}
}

func TestGameService_ApplyAction(t *testing.T) {
	gameRepo := repository.NewGameRepository()
	gameService := service.NewGameService(gameRepo, events.NewInMemoryEventBus(), nil)

	// Create a game and add a player
	settings := model.GameSettings{MaxPlayers: 4}
	game, err := gameService.CreateGame(settings)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	game, err = gameService.JoinGame(game.ID, "Player1")
	if err != nil {
		t.Fatalf("Failed to join game: %v", err)
	}

	playerID := game.Players[0].ID

	tests := []struct {
		name          string
		gameID        string
		playerID      string
		actionRequest interface{}
		wantErr       bool
	}{
		{
			name:     "invalid action request type",
			gameID:   game.ID,
			playerID: playerID,
			actionRequest: "invalid-request",
			wantErr: true,
		},
		{
			name:     "non-existent game",
			gameID:   "non-existent",
			playerID: playerID,
			actionRequest: dto.ActionStartGameRequest{
				Type: dto.ActionTypeStartGame,
			},
			wantErr: true,
		},
		{
			name:     "non-existent player",
			gameID:   game.ID,
			playerID: "non-existent",
			actionRequest: dto.ActionStartGameRequest{
				Type: dto.ActionTypeStartGame,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedGame, err := gameService.ApplyAction(tt.gameID, tt.playerID, tt.actionRequest)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyAction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && updatedGame == nil {
				t.Error("Expected updated game to be non-nil")
			}
		})
	}
}
