package websocket_test

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket"
	model "terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
)

// Helper functions for creating pointers
func stringPtr(s string) *string { return &s }
func intPtr(i int) *int         { return &i }

// Helper function to create GameService with all dependencies for testing
func createTestGameService() *service.GameService {
	gameRepo := repository.NewGameRepository()
	cardSelectionRepo := repository.NewCardSelectionRepository()
	eventBus := events.NewInMemoryEventBus()
	eventRepository := events.NewEventRepository(eventBus)
	playerService := service.NewPlayerService(gameRepo, eventBus, eventRepository)
	
	return service.NewGameService(gameRepo, cardSelectionRepo, eventBus, eventRepository, nil, playerService)
}

// mockClient implements basic client functionality for testing
type mockClient struct {
	ID       string
	PlayerID string
	GameID   string
	send     chan []byte
	hub      *websocket.Hub
	messages []dto.WebSocketMessage
	mutex    sync.Mutex
}

func newMockClient(hub *websocket.Hub) *mockClient {
	return &mockClient{
		ID:       "test-client-" + time.Now().Format("20060102150405"),
		send:     make(chan []byte, 256),
		hub:      hub,
		messages: make([]dto.WebSocketMessage, 0),
	}
}

func (c *mockClient) SetPlayerInfo(playerID, gameID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.PlayerID = playerID
	c.GameID = gameID
}

func (c *mockClient) GetPlayerInfo() (string, string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.PlayerID, c.GameID
}

func (c *mockClient) SendMessage(msg *dto.WebSocketMessage) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.messages = append(c.messages, *msg)
}

func (c *mockClient) GetMessages() []dto.WebSocketMessage {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return append([]dto.WebSocketMessage(nil), c.messages...)
}

func (c *mockClient) ClearMessages() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.messages = make([]dto.WebSocketMessage, 0)
}

func TestNewHub(t *testing.T) {
	gameService := createTestGameService()
	hub := websocket.NewHub(gameService)

	if hub == nil {
		t.Fatal("Expected hub to be non-nil")
	}
}

func TestHub_BroadcastToGame(t *testing.T) {
	gameService := createTestGameService()

	// Create a test game
	settings := model.GameSettings{MaxPlayers: 4}
	game, err := gameService.CreateGame(settings)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	hub := websocket.NewHub(gameService)

	// Test broadcasting to non-existent game
	message := &dto.WebSocketMessage{
		Type: dto.MessageTypeGameUpdated,
		Payload: dto.GameUpdatedPayload{
			Game: dto.ToGameDto(game),
		},
		GameID: "non-existent-game",
	}

	// This should not panic
	hub.BroadcastToGame("non-existent-game", message)
}

func TestHub_PayloadParsing(t *testing.T) {
	tests := []struct {
		name    string
		payload interface{}
		target  interface{}
		wantErr bool
	}{
		{
			name: "valid payload parsing",
			payload: map[string]interface{}{
				"gameId":     "test-game",
				"playerName": "TestPlayer",
			},
			target:  &dto.PlayerConnectPayload{},
			wantErr: false,
		},
		{
			name: "valid action payload parsing",
			payload: map[string]interface{}{
				"actionRequest": map[string]interface{}{
					"type": "start-game",
				},
			},
			target:  &dto.PlayActionPayload{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert payload to JSON and back to simulate WebSocket message parsing
			data, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal test payload: %v", err)
			}

			err = json.Unmarshal(data, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the parsing worked correctly
				switch target := tt.target.(type) {
				case *dto.PlayerConnectPayload:
					if target.GameID != "test-game" || target.PlayerName != "TestPlayer" {
						t.Errorf("PlayerConnectPayload not parsed correctly: %+v", target)
					}
				case *dto.PlayActionPayload:
					if requestMap, ok := target.ActionRequest.(map[string]interface{}); ok {
						if actionType, exists := requestMap["type"]; !exists || actionType != "start-game" {
							t.Errorf("PlayActionPayload action type not parsed correctly: %+v", target)
						}
					} else {
						t.Errorf("PlayActionPayload ActionRequest not parsed as map: %+v", target)
					}
				}
			}
		})
	}
}
