package award

import (
	"fmt"
)

// AwardRegistry provides lookup functionality for award definitions
type AwardRegistry interface {
	GetByID(awardID string) (*AwardDefinition, error)
	GetAll() []AwardDefinition
}

// InMemoryAwardRegistry implements AwardRegistry with an in-memory map
type InMemoryAwardRegistry struct {
	awards map[string]AwardDefinition
	order  []string
}

// NewInMemoryAwardRegistry creates a new registry from a slice of definitions
func NewInMemoryAwardRegistry(awardList []AwardDefinition) *InMemoryAwardRegistry {
	awardMap := make(map[string]AwardDefinition, len(awardList))
	order := make([]string, 0, len(awardList))
	for _, a := range awardList {
		awardMap[a.ID] = a
		order = append(order, a.ID)
	}
	return &InMemoryAwardRegistry{awards: awardMap, order: order}
}

// GetByID retrieves an award definition by ID
func (r *InMemoryAwardRegistry) GetByID(awardID string) (*AwardDefinition, error) {
	a, exists := r.awards[awardID]
	if !exists {
		return nil, fmt.Errorf("award not found: %s", awardID)
	}
	return &a, nil
}

// GetAll returns all award definitions in their original JSON order
func (r *InMemoryAwardRegistry) GetAll() []AwardDefinition {
	result := make([]AwardDefinition, 0, len(r.order))
	for _, id := range r.order {
		result = append(result, r.awards[id])
	}
	return result
}
