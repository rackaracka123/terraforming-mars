package corporation

import (
	"sync"

	"terraforming-mars-backend/internal/session/game/card"
)

// Corporation manages player corporation selection and the corporation card.
// Thread-safe with its own mutex.
type Corporation struct {
	mu            sync.RWMutex
	corporation   *card.Card // The corporation card (nil if not selected)
	corporationID string     // Corporation ID for quick reference
}

// NewCorporation creates a new Corporation component with no corporation selected.
func NewCorporation() *Corporation {
	return &Corporation{
		corporation:   nil,
		corporationID: "",
	}
}

// ==================== Getters ====================

// Card returns the corporation card.
// Returns nil if no corporation has been selected.
func (c *Corporation) Card() *card.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.corporation
}

// ID returns the corporation ID.
// Returns empty string if no corporation has been selected.
func (c *Corporation) ID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.corporationID
}

// ==================== Setters ====================

// SetID sets the corporation ID.
func (c *Corporation) SetID(corporationID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.corporationID = corporationID
}

// SetCard sets the corporation card and updates the ID.
func (c *Corporation) SetCard(corporation card.Card) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.corporation = &corporation
	c.corporationID = corporation.ID
}

// ==================== Utilities ====================

// HasCorporation returns true if a corporation has been selected.
func (c *Corporation) HasCorporation() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.corporation != nil
}

// DeepCopy creates a deep copy of the Corporation component.
func (c *Corporation) DeepCopy() *Corporation {
	if c == nil {
		return nil
	}

	var corpCopy *card.Card
	if c.corporation != nil {
		// Create a copy of the corporation card
		corpValue := *c.corporation
		corpCopy = &corpValue
	}

	return &Corporation{
		corporation:   corpCopy,
		corporationID: c.corporationID,
	}
}
