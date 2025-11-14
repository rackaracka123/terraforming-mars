package tiles

import (
	"context"
	"fmt"
	"sync"
)

// PendingTileSelection represents a pending tile placement action
type PendingTileSelection struct {
	TileType       string   `json:"tileType" ts:"string"`         // "city", "greenery", "ocean"
	AvailableHexes []string `json:"availableHexes" ts:"string[]"` // Backend-calculated valid hex coordinates
	Source         string   `json:"source" ts:"string"`           // What triggered this selection (card ID, standard project, etc.)
}

// PendingTileSelectionQueue represents a queue of tile placements to be made
type PendingTileSelectionQueue struct {
	Items  []string `json:"items" ts:"string[]"` // Queue of tile types: ["city", "city", "ocean"]
	Source string   `json:"source" ts:"string"`  // Card ID that triggered all placements
}

// TileQueueRepository manages pending tile selections for a player
type TileQueueRepository interface {
	// Get queue state
	GetQueue(ctx context.Context) (*PendingTileSelectionQueue, error)
	GetPendingSelection(ctx context.Context) (*PendingTileSelection, error)

	// Bulk queue operations
	SetQueue(ctx context.Context, queue *PendingTileSelectionQueue) error

	// Granular queue operations
	AddToQueue(ctx context.Context, tileType string) error
	PopFromQueue(ctx context.Context) (string, error)
	ClearQueue(ctx context.Context) error
	GetQueueLength(ctx context.Context) (int, error)

	// Pending selection operations
	SetPendingSelection(ctx context.Context, selection *PendingTileSelection) error
	ClearPendingSelection(ctx context.Context) error
}

// TileQueueRepositoryImpl implements independent in-memory storage for tile queue
type TileQueueRepositoryImpl struct {
	mu               sync.RWMutex
	queue            *PendingTileSelectionQueue
	pendingSelection *PendingTileSelection
}

// NewTileQueueRepository creates a new independent tile queue repository with initial state
func NewTileQueueRepository(initialQueue *PendingTileSelectionQueue, initialSelection *PendingTileSelection) TileQueueRepository {
	var queueCopy *PendingTileSelectionQueue
	if initialQueue != nil {
		copy := *initialQueue
		copy.Items = append([]string{}, initialQueue.Items...)
		queueCopy = &copy
	}

	var selectionCopy *PendingTileSelection
	if initialSelection != nil {
		copy := *initialSelection
		copy.AvailableHexes = append([]string{}, initialSelection.AvailableHexes...)
		selectionCopy = &copy
	}

	return &TileQueueRepositoryImpl{
		queue:            queueCopy,
		pendingSelection: selectionCopy,
	}
}

// GetQueue retrieves the current queue
func (r *TileQueueRepositoryImpl) GetQueue(ctx context.Context) (*PendingTileSelectionQueue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.queue == nil {
		return nil, nil
	}

	// Return a copy
	queueCopy := *r.queue
	queueCopy.Items = append([]string{}, r.queue.Items...)
	return &queueCopy, nil
}

// SetQueue sets the queue to specific values (for bulk updates)
func (r *TileQueueRepositoryImpl) SetQueue(ctx context.Context, queue *PendingTileSelectionQueue) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if queue == nil {
		r.queue = nil
		return nil
	}

	// Make a copy
	queueCopy := *queue
	queueCopy.Items = append([]string{}, queue.Items...)
	r.queue = &queueCopy

	return nil
}

// GetPendingSelection retrieves the current pending selection
func (r *TileQueueRepositoryImpl) GetPendingSelection(ctx context.Context) (*PendingTileSelection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.pendingSelection == nil {
		return nil, nil
	}

	// Return a copy
	selectionCopy := *r.pendingSelection
	selectionCopy.AvailableHexes = append([]string{}, r.pendingSelection.AvailableHexes...)
	return &selectionCopy, nil
}

// AddToQueue adds a tile type to the queue
func (r *TileQueueRepositoryImpl) AddToQueue(ctx context.Context, tileType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.queue == nil {
		r.queue = &PendingTileSelectionQueue{
			Items:  []string{tileType},
			Source: "", // Source will be set when processing
		}
	} else {
		r.queue.Items = append(r.queue.Items, tileType)
	}

	return nil
}

// PopFromQueue removes and returns the first tile type from the queue
func (r *TileQueueRepositoryImpl) PopFromQueue(ctx context.Context) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.queue == nil || len(r.queue.Items) == 0 {
		return "", fmt.Errorf("queue is empty")
	}

	tileType := r.queue.Items[0]
	r.queue.Items = r.queue.Items[1:]

	// If queue is empty, clear it
	if len(r.queue.Items) == 0 {
		r.queue = nil
	}

	return tileType, nil
}

// ClearQueue removes all items from the queue
func (r *TileQueueRepositoryImpl) ClearQueue(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.queue = nil
	return nil
}

// GetQueueLength returns the number of items in the queue
func (r *TileQueueRepositoryImpl) GetQueueLength(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.queue == nil {
		return 0, nil
	}

	return len(r.queue.Items), nil
}

// SetPendingSelection sets the current pending tile selection
func (r *TileQueueRepositoryImpl) SetPendingSelection(ctx context.Context, selection *PendingTileSelection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if selection == nil {
		r.pendingSelection = nil
		return nil
	}

	// Make a copy
	selectionCopy := *selection
	selectionCopy.AvailableHexes = append([]string{}, selection.AvailableHexes...)
	r.pendingSelection = &selectionCopy

	return nil
}

// ClearPendingSelection clears the current pending selection
func (r *TileQueueRepositoryImpl) ClearPendingSelection(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pendingSelection = nil
	return nil
}
