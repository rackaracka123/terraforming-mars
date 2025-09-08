package utils

import (
	"encoding/json"
	"fmt"
)

// MessageParser provides unified JSON parsing for WebSocket messages
type MessageParser struct{}

// NewMessageParser creates a new message parser
func NewMessageParser() *MessageParser {
	return &MessageParser{}
}

// ParsePayload parses a WebSocket message payload into the given destination
// This unifies the previously duplicate parseMessagePayload and parseActionRequest functions
func (mp *MessageParser) ParsePayload(payload interface{}, dest interface{}) error {
	// Convert payload to JSON bytes and then unmarshal to the destination
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, dest); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}

// ParseAction parses an action request and extracts the action type
func (mp *MessageParser) ParseAction(actionRequest interface{}) (actionType string, actionMap map[string]interface{}, err error) {
	// Parse action request into a map to extract action type
	actionBytes, err := json.Marshal(actionRequest)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal action request: %w", err)
	}

	if err := json.Unmarshal(actionBytes, &actionMap); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal action request: %w", err)
	}

	actionType, ok := actionMap["type"].(string)
	if !ok {
		return "", nil, fmt.Errorf("action type not found or invalid")
	}

	return actionType, actionMap, nil
}

// ParseTypedAction parses an action request into a specific type
func (mp *MessageParser) ParseTypedAction(actionRequest interface{}, dest interface{}) error {
	return mp.ParsePayload(actionRequest, dest)
}