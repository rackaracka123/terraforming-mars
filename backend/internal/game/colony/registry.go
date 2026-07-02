package colony

import (
	"fmt"
)

// ColonyRegistry provides lookup functionality for colony tile definitions
type ColonyRegistry interface {
	GetByID(colonyID string) (*ColonyDefinition, error)
	GetAll() []ColonyDefinition
}

// InMemoryColonyRegistry implements ColonyRegistry with an in-memory map
type InMemoryColonyRegistry struct {
	colonies map[string]ColonyDefinition
	order    []string // preserves load (JSON) order so GetAll is deterministic
}

// NewInMemoryColonyRegistry creates a new colony registry from a slice of definitions
func NewInMemoryColonyRegistry(colonyList []ColonyDefinition) *InMemoryColonyRegistry {
	colonyMap := make(map[string]ColonyDefinition, len(colonyList))
	order := make([]string, 0, len(colonyList))
	for _, c := range colonyList {
		if _, exists := colonyMap[c.ID]; !exists {
			order = append(order, c.ID)
		}
		colonyMap[c.ID] = c
	}
	return &InMemoryColonyRegistry{colonies: colonyMap, order: order}
}

// GetByID retrieves a colony tile definition by ID
func (r *InMemoryColonyRegistry) GetByID(colonyID string) (*ColonyDefinition, error) {
	c, exists := r.colonies[colonyID]
	if !exists {
		return nil, fmt.Errorf("colony not found: %s", colonyID)
	}
	return &c, nil
}

// GetAll returns all colony tile definitions in their original load (JSON) order
// so that seeded colony selection is reproducible for a given seed.
func (r *InMemoryColonyRegistry) GetAll() []ColonyDefinition {
	result := make([]ColonyDefinition, 0, len(r.order))
	for _, id := range r.order {
		result = append(result, r.colonies[id])
	}
	return result
}
