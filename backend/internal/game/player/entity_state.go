package player

import "time"

// EntityState holds calculated state for any entity (card, action, project).
// This generic structure eliminates redundant boolean flags and works across all entity types.
// The single source of truth is the Errors slice - availability is computed from it.
type EntityState struct {
	// Single source of truth - Errors determine availability
	Errors []StateError

	// Multi-resource cost map (empty map means no cost)
	// Keys are resource types like "credits", "plants", "heat"
	// Values are the amounts required
	Cost map[string]int

	// Minimal entity-specific data (prefer typed fields over metadata when predictable)
	Metadata map[string]interface{}

	// Calculation timestamp
	LastCalculated time.Time
}

// Available returns true if there are no errors (computed, not stored).
// This prevents contradictory state between availability flags and error lists.
func (e EntityState) Available() bool {
	return len(e.Errors) == 0
}

// StateError represents a specific reason why an entity is unavailable.
// Errors are categorized for UI filtering and display.
type StateError struct {
	// Error code for programmatic handling (e.g., ErrorCodeInsufficientCredits)
	Code StateErrorCode

	// Category for grouping errors (e.g., ErrorCategoryPhase, ErrorCategoryCost)
	Category StateErrorCategory

	// Human-readable error message
	Message string
}
