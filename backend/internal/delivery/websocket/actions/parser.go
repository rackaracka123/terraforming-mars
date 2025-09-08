package actions

import (
	"encoding/json"
	"fmt"
)

// MessageParser provides JSON parsing for action requests
type MessageParser struct{}

// NewMessageParser creates a new message parser
func NewMessageParser() *MessageParser {
	return &MessageParser{}
}

// ParsePayload parses an action payload into the given destination
func (mp *MessageParser) ParsePayload(payload interface{}, dest interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, dest); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}