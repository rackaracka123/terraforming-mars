package player

import (
	"sync"

	"terraforming-mars-backend/internal/game/shared"
)

type GenerationalEvents struct {
	mu     sync.RWMutex
	counts map[shared.GenerationalEvent]int
}

func newGenerationalEvents() *GenerationalEvents {
	return &GenerationalEvents{
		counts: make(map[shared.GenerationalEvent]int),
	}
}

func (ge *GenerationalEvents) Increment(event shared.GenerationalEvent) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.counts[event]++
}

func (ge *GenerationalEvents) GetCount(event shared.GenerationalEvent) int {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.counts[event]
}

func (ge *GenerationalEvents) GetAll() []shared.PlayerGenerationalEventEntry {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	entries := make([]shared.PlayerGenerationalEventEntry, 0, len(ge.counts))
	for event, count := range ge.counts {
		entries = append(entries, shared.PlayerGenerationalEventEntry{
			Event: event,
			Count: count,
		})
	}
	return entries
}

func (ge *GenerationalEvents) Clear() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.counts = make(map[shared.GenerationalEvent]int)
}
