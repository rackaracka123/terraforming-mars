package common

import (
	"context"
)

// UseCase represents a generic use case interface
type UseCase[TRequest, TResponse any] interface {
	Execute(ctx context.Context, request TRequest) (TResponse, error)
}

// ActionCost represents the cost of an action in multiple resource types
type ActionCost struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// ActionValidator validates common requirements for player actions
type ActionValidator interface {
	// ValidatePlayerAction checks if a player can perform an action
	// Validates: player's turn, remaining actions, and sufficient resources
	ValidatePlayerAction(ctx context.Context, gameID, playerID string, cost ActionCost) error
}

// ActionRequest represents a common action request structure
type ActionRequest struct {
	GameID   string `json:"gameId" ts:"string"`
	PlayerID string `json:"playerId" ts:"string"`
}

// ActionResponse represents a common action response structure
type ActionResponse struct {
	Success bool   `json:"success" ts:"boolean"`
	Message string `json:"message" ts:"string"`
}
