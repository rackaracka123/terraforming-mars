package websocket_test

import (
	"encoding/json"
	"reflect"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/model"
	"testing"
)

func TestMessageType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		msgType  websocket.MessageType
		expected string
	}{
		{
			name:     "player connect message type",
			msgType:  websocket.MessageTypePlayerConnect,
			expected: "player-connect",
		},
		{
			name:     "play action message type",
			msgType:  websocket.MessageTypePlayAction,
			expected: "play-action",
		},
		{
			name:     "game updated message type",
			msgType:  websocket.MessageTypeGameUpdated,
			expected: "game-updated",
		},
		{
			name:     "player connected message type",
			msgType:  websocket.MessageTypePlayerConnected,
			expected: "player-connected",
		},
		{
			name:     "error message type",
			msgType:  websocket.MessageTypeError,
			expected: "error",
		},
		{
			name:     "full state message type",
			msgType:  websocket.MessageTypeFullState,
			expected: "full-state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.msgType) != tt.expected {
				t.Errorf("Expected message type %s, got %s", tt.expected, string(tt.msgType))
			}
		})
	}
}

func TestWebSocketMessage_JSONSerialization(t *testing.T) {
	tests := []struct {
		name    string
		message websocket.WebSocketMessage
	}{
		{
			name: "player connect message",
			message: websocket.WebSocketMessage{
				Type: websocket.MessageTypePlayerConnect,
				Payload: websocket.PlayerConnectPayload{
					PlayerName: "TestPlayer",
					GameID:     "test-game-123",
				},
				GameID: "test-game-123",
			},
		},
		{
			name: "play action message",
			message: websocket.WebSocketMessage{
				Type: websocket.MessageTypePlayAction,
				Payload: websocket.PlayActionPayload{
					Action: "skip-action",
					Data:   map[string]interface{}{"test": "value"},
				},
			},
		},
		{
			name: "error message",
			message: websocket.WebSocketMessage{
				Type: websocket.MessageTypeError,
				Payload: websocket.ErrorPayload{
					Message: "Test error message",
					Code:    "TEST_ERROR",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize to JSON
			data, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatalf("Failed to marshal message: %v", err)
			}

			// Deserialize from JSON
			var deserializedMessage websocket.WebSocketMessage
			err = json.Unmarshal(data, &deserializedMessage)
			if err != nil {
				t.Fatalf("Failed to unmarshal message: %v", err)
			}

			// Check that message type is preserved
			if deserializedMessage.Type != tt.message.Type {
				t.Errorf("Message type not preserved: expected %s, got %s",
					tt.message.Type, deserializedMessage.Type)
			}

			// Check that GameID is preserved
			if deserializedMessage.GameID != tt.message.GameID {
				t.Errorf("GameID not preserved: expected %s, got %s",
					tt.message.GameID, deserializedMessage.GameID)
			}

			// Payload comparison is more complex due to interface{} type
			// We'll verify it exists and can be re-parsed
			if deserializedMessage.Payload == nil && tt.message.Payload != nil {
				t.Error("Payload was lost during serialization")
			}
		})
	}
}

func TestPlayerConnectPayload_JSONSerialization(t *testing.T) {
	payload := websocket.PlayerConnectPayload{
		PlayerName: "TestPlayer",
		GameID:     "test-game-123",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal PlayerConnectPayload: %v", err)
	}

	var deserializedPayload websocket.PlayerConnectPayload
	err = json.Unmarshal(data, &deserializedPayload)
	if err != nil {
		t.Fatalf("Failed to unmarshal PlayerConnectPayload: %v", err)
	}

	if !reflect.DeepEqual(payload, deserializedPayload) {
		t.Errorf("PlayerConnectPayload not preserved: expected %+v, got %+v",
			payload, deserializedPayload)
	}
}

func TestPlayActionPayload_JSONSerialization(t *testing.T) {
	tests := []struct {
		name    string
		payload websocket.PlayActionPayload
	}{
		{
			name: "action without data",
			payload: websocket.PlayActionPayload{
				Action: "skip-action",
				Data:   nil,
			},
		},
		{
			name: "action with data",
			payload: websocket.PlayActionPayload{
				Action: "raise-temperature",
				Data: map[string]interface{}{
					"heatAmount": 8.0,
				},
			},
		},
		{
			name: "corporation selection action",
			payload: websocket.PlayActionPayload{
				Action: "select-corporation",
				Data: map[string]interface{}{
					"corporationName": "TestCorp",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal PlayActionPayload: %v", err)
			}

			var deserializedPayload websocket.PlayActionPayload
			err = json.Unmarshal(data, &deserializedPayload)
			if err != nil {
				t.Fatalf("Failed to unmarshal PlayActionPayload: %v", err)
			}

			if deserializedPayload.Action != tt.payload.Action {
				t.Errorf("Action not preserved: expected %s, got %s",
					tt.payload.Action, deserializedPayload.Action)
			}

			// For data comparison, we need to handle the interface{} type carefully
			if tt.payload.Data == nil && deserializedPayload.Data != nil {
				t.Error("Expected nil data, but got non-nil")
			}
			if tt.payload.Data != nil && deserializedPayload.Data == nil {
				t.Error("Expected non-nil data, but got nil")
			}
		})
	}
}

func TestGameUpdatedPayload_JSONSerialization(t *testing.T) {
	// Create a test game
	game := &domain.Game{
		ID: "test-game-123",
		Settings: domain.GameSettings{
			MaxPlayers: 4,
		},
		Status: domain.GameStatusWaiting,
		Players: []domain.Player{
			{
				ID:   "player-1",
				Name: "Player 1",
				Resources: domain.Resources{
					Credits: 10,
				},
				Production: domain.Production{
					Credits: 1,
				},
				TerraformRating: 20,
				IsActive:        true,
				PlayedCards:     []string{},
			},
		},
		GlobalParameters: domain.GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		},
	}

	payload := websocket.GameUpdatedPayload{
		Game: dto.ToGameDto(game),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal GameUpdatedPayload: %v", err)
	}

	var deserializedPayload websocket.GameUpdatedPayload
	err = json.Unmarshal(data, &deserializedPayload)
	if err != nil {
		t.Fatalf("Failed to unmarshal GameUpdatedPayload: %v", err)
	}

	// Verify game data is preserved
	if deserializedPayload.Game.ID == "" {
		t.Fatal("Game ID is empty after deserialization")
	}

	if deserializedPayload.Game.ID != game.ID {
		t.Errorf("Game ID not preserved: expected %s, got %s",
			game.ID, deserializedPayload.Game.ID)
	}

	// Note: Game struct doesn't have Name field in domain model

	if len(deserializedPayload.Game.Players) != len(game.Players) {
		t.Errorf("Player count not preserved: expected %d, got %d",
			len(game.Players), len(deserializedPayload.Game.Players))
	}
}

func TestPlayerConnectedPayload_JSONSerialization(t *testing.T) {
	payload := websocket.PlayerConnectedPayload{
		PlayerID:   "player-123",
		PlayerName: "TestPlayer",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal PlayerConnectedPayload: %v", err)
	}

	var deserializedPayload websocket.PlayerConnectedPayload
	err = json.Unmarshal(data, &deserializedPayload)
	if err != nil {
		t.Fatalf("Failed to unmarshal PlayerConnectedPayload: %v", err)
	}

	if !reflect.DeepEqual(payload, deserializedPayload) {
		t.Errorf("PlayerConnectedPayload not preserved: expected %+v, got %+v",
			payload, deserializedPayload)
	}
}

func TestErrorPayload_JSONSerialization(t *testing.T) {
	tests := []struct {
		name    string
		payload websocket.ErrorPayload
	}{
		{
			name: "error with code",
			payload: websocket.ErrorPayload{
				Message: "Test error message",
				Code:    "TEST_ERROR",
			},
		},
		{
			name: "error without code",
			payload: websocket.ErrorPayload{
				Message: "Test error message without code",
				Code:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal ErrorPayload: %v", err)
			}

			var deserializedPayload websocket.ErrorPayload
			err = json.Unmarshal(data, &deserializedPayload)
			if err != nil {
				t.Fatalf("Failed to unmarshal ErrorPayload: %v", err)
			}

			if !reflect.DeepEqual(tt.payload, deserializedPayload) {
				t.Errorf("ErrorPayload not preserved: expected %+v, got %+v",
					tt.payload, deserializedPayload)
			}
		})
	}
}

func TestFullStatePayload_JSONSerialization(t *testing.T) {
	// Create a test game
	game := &domain.Game{
		ID: "test-game-123",
		Settings: domain.GameSettings{
			MaxPlayers: 2,
		},
		Status: domain.GameStatusActive,
		Players: []domain.Player{
			{
				ID:              "player-1",
				Name:            "Player 1",
				Corporation:     "TestCorp",
				Resources:       domain.Resources{Credits: 25},
				Production:      domain.Production{Credits: 2},
				TerraformRating: 22,
				IsActive:        true,
				PlayedCards:     []string{"card-1", "card-2"},
			},
		},
		GlobalParameters: domain.GlobalParameters{
			Temperature: -26,
			Oxygen:      2,
			Oceans:      1,
		},
	}

	payload := websocket.FullStatePayload{
		Game:     dto.ToGameDto(game),
		PlayerID: "player-1",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal FullStatePayload: %v", err)
	}

	var deserializedPayload websocket.FullStatePayload
	err = json.Unmarshal(data, &deserializedPayload)
	if err != nil {
		t.Fatalf("Failed to unmarshal FullStatePayload: %v", err)
	}

	// Verify player ID is preserved
	if deserializedPayload.PlayerID != payload.PlayerID {
		t.Errorf("PlayerID not preserved: expected %s, got %s",
			payload.PlayerID, deserializedPayload.PlayerID)
	}

	// Verify game data is preserved
	if deserializedPayload.Game.ID == "" {
		t.Fatal("Game ID is empty after deserialization")
	}

	if deserializedPayload.Game.ID != game.ID {
		t.Errorf("Game ID not preserved: expected %s, got %s",
			game.ID, deserializedPayload.Game.ID)
	}

	// Verify complex nested data is preserved
	if len(deserializedPayload.Game.Players) != 1 {
		t.Fatalf("Expected 1 player, got %d", len(deserializedPayload.Game.Players))
	}

	player := deserializedPayload.Game.Players[0]
	if player.Corporation != "TestCorp" {
		t.Errorf("Player corporation not preserved: expected %s, got %s",
			"TestCorp", player.Corporation)
	}

	if len(player.PlayedCards) != 2 {
		t.Errorf("Played cards not preserved: expected 2 cards, got %d",
			len(player.PlayedCards))
	}
}

func TestMessage_PayloadParsing(t *testing.T) {
	// Test that a generic WebSocket message can have its payload properly parsed
	// based on the message type
	tests := []struct {
		name        string
		messageJSON string
		messageType websocket.MessageType
	}{
		{
			name: "player connect message parsing",
			messageJSON: `{
				"type": "player-connect",
				"payload": {
					"playerName": "TestPlayer",
					"gameId": "test-game-123"
				},
				"gameId": "test-game-123"
			}`,
			messageType: websocket.MessageTypePlayerConnect,
		},
		{
			name: "play action message parsing",
			messageJSON: `{
				"type": "play-action",
				"payload": {
					"action": "raise-temperature",
					"data": {
						"heatAmount": 16
					}
				}
			}`,
			messageType: websocket.MessageTypePlayAction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var message websocket.WebSocketMessage
			err := json.Unmarshal([]byte(tt.messageJSON), &message)
			if err != nil {
				t.Fatalf("Failed to unmarshal message JSON: %v", err)
			}

			if message.Type != tt.messageType {
				t.Errorf("Message type not parsed correctly: expected %s, got %s",
					tt.messageType, message.Type)
			}

			if message.Payload == nil {
				t.Error("Payload is nil after parsing")
			}

			// Test that payload can be re-marshaled and parsed into specific types
			payloadData, err := json.Marshal(message.Payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			switch tt.messageType {
			case websocket.MessageTypePlayerConnect:
				var payload websocket.PlayerConnectPayload
				err = json.Unmarshal(payloadData, &payload)
				if err != nil {
					t.Fatalf("Failed to parse PlayerConnectPayload: %v", err)
				}
				if payload.PlayerName != "TestPlayer" {
					t.Errorf("PlayerName not parsed correctly: got %s", payload.PlayerName)
				}

			case websocket.MessageTypePlayAction:
				var payload websocket.PlayActionPayload
				err = json.Unmarshal(payloadData, &payload)
				if err != nil {
					t.Fatalf("Failed to parse PlayActionPayload: %v", err)
				}
				if payload.Action != "raise-temperature" {
					t.Errorf("Action not parsed correctly: got %s", payload.Action)
				}
			}
		})
	}
}
