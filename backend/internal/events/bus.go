package events

import (
	"context"
	"errors"
	"sync"
	"terraforming-mars-backend/internal/logger"
	"time"

	"go.uber.org/zap"
)

var (
	// ErrEventBusClosed is returned when trying to use a closed event bus
	ErrEventBusClosed = errors.New("event bus is closed")
)

// EventListener represents a function that handles an event
type EventListener func(ctx context.Context, event Event) error

// EventBus defines the interface for event publishing and subscription
type EventBus interface {
	// Subscribe registers a listener for events of the specified type
	Subscribe(eventType string, listener EventListener)
	// Publish sends an event to all registered listeners for its type
	Publish(ctx context.Context, event Event) error
	// Unsubscribe removes a listener (if needed for testing)
	Unsubscribe(eventType string, listener EventListener)
	// Close shuts down the event bus and its worker pool
	Close() error
}

// eventJob represents a job to be processed by the worker pool
type eventJob struct {
	ctx      context.Context
	event    Event
	listener EventListener
}

// InMemoryEventBus implements EventBus using in-memory subscription storage with worker pool
type InMemoryEventBus struct {
	listeners map[string][]EventListener
	mutex     sync.RWMutex
	jobQueue  chan eventJob
	workers   int
	workerWg  sync.WaitGroup
	closeOnce sync.Once
	closed    chan struct{}
	workerSem chan struct{} // Semaphore to limit concurrent workers
}

// NewInMemoryEventBus creates a new in-memory event bus with worker pool
func NewInMemoryEventBus() *InMemoryEventBus {
	const (
		defaultWorkers = 10
		bufferSize     = 1000
	)

	bus := &InMemoryEventBus{
		listeners: make(map[string][]EventListener),
		jobQueue:  make(chan eventJob, bufferSize),
		workers:   defaultWorkers,
		closed:    make(chan struct{}),
		workerSem: make(chan struct{}, defaultWorkers),
	}

	// Start worker pool
	bus.startWorkers()

	return bus
}

// NewInMemoryEventBusWithWorkers creates a new event bus with specified worker count
func NewInMemoryEventBusWithWorkers(workerCount, bufferSize int) *InMemoryEventBus {
	if workerCount <= 0 {
		workerCount = 10
	}
	if bufferSize <= 0 {
		bufferSize = 1000
	}

	bus := &InMemoryEventBus{
		listeners: make(map[string][]EventListener),
		jobQueue:  make(chan eventJob, bufferSize),
		workers:   workerCount,
		closed:    make(chan struct{}),
		workerSem: make(chan struct{}, workerCount),
	}

	// Start worker pool
	bus.startWorkers()

	return bus
}

// startWorkers initializes the worker pool
func (bus *InMemoryEventBus) startWorkers() {
	logger.Info("🔥 Starting event bus worker pool", zap.Int("workers", bus.workers))

	for i := 0; i < bus.workers; i++ {
		bus.workerWg.Add(1)
		go bus.worker(i)
	}
}

// worker processes event jobs from the queue
func (bus *InMemoryEventBus) worker(id int) {
	defer bus.workerWg.Done()
	log := logger.WithContext(zap.Int("worker_id", id))

	log.Debug("👷 Event worker started")
	defer log.Debug("👷 Event worker stopped")

	for {
		select {
		case <-bus.closed:
			return
		case job := <-bus.jobQueue:
			// Acquire semaphore to limit concurrency
			bus.workerSem <- struct{}{}

			// Process the event with timeout and proper error handling
			func() {
				defer func() {
					<-bus.workerSem // Release semaphore
					if r := recover(); r != nil {
						log.Error("🚨 Event listener panicked",
							zap.Any("panic", r),
							zap.String("event_type", job.event.GetType()))
					}
				}()

				// Create timeout context for event processing
				ctx, cancel := context.WithTimeout(job.ctx, 30*time.Second)
				defer cancel()

				if err := job.listener(ctx, job.event); err != nil {
					log.Error("❌ Event listener failed",
						zap.String("event_type", job.event.GetType()),
						zap.String("game_id", job.event.GetGameID()),
						zap.Error(err))
				}
			}()
		}
	}
}

// Subscribe registers a listener for events of the specified type
func (bus *InMemoryEventBus) Subscribe(eventType string, listener EventListener) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	if bus.listeners[eventType] == nil {
		bus.listeners[eventType] = make([]EventListener, 0)
	}

	bus.listeners[eventType] = append(bus.listeners[eventType], listener)

	logger.Info("Event listener registered",
		zap.String("event_type", eventType),
		zap.Int("listener_count", len(bus.listeners[eventType])),
	)
}

// Publish sends an event to all registered listeners for its type
func (bus *InMemoryEventBus) Publish(ctx context.Context, event Event) error {
	// Check if bus is closed
	select {
	case <-bus.closed:
		return ErrEventBusClosed
	default:
	}

	bus.mutex.RLock()
	listeners := bus.listeners[event.GetType()]
	bus.mutex.RUnlock()

	log := logger.WithGameContext(event.GetGameID(), "")

	if len(listeners) == 0 {
		log.Debug("📭 No listeners registered for event type",
			zap.String("event_type", event.GetType()),
		)
		return nil
	}

	log.Info("📤 Publishing event to worker pool",
		zap.String("event_type", event.GetType()),
		zap.Int("listener_count", len(listeners)),
	)

	// Queue all listener jobs for worker pool processing
	jobsQueued := 0
	for _, listener := range listeners {
		job := eventJob{
			ctx:      ctx,
			event:    event,
			listener: listener,
		}

		select {
		case bus.jobQueue <- job:
			jobsQueued++
		case <-ctx.Done():
			log.Warn("⏰ Context cancelled while queueing event jobs",
				zap.String("event_type", event.GetType()),
				zap.Int("jobs_queued", jobsQueued),
				zap.Int("total_listeners", len(listeners)))
			return ctx.Err()
		case <-bus.closed:
			log.Warn("🚫 Event bus closed while queueing jobs",
				zap.String("event_type", event.GetType()),
				zap.Int("jobs_queued", jobsQueued))
			return ErrEventBusClosed
		default:
			// Job queue is full, log warning but continue
			log.Warn("⚠️ Event job queue full, dropping event",
				zap.String("event_type", event.GetType()),
				zap.String("game_id", event.GetGameID()))
		}
	}

	log.Info("✅ Event jobs queued for processing",
		zap.String("event_type", event.GetType()),
		zap.Int("jobs_queued", jobsQueued),
		zap.Int("total_listeners", len(listeners)),
	)

	return nil
}

// Unsubscribe removes a listener from the event type (used mainly for testing)
func (bus *InMemoryEventBus) Unsubscribe(eventType string, listener EventListener) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	listeners := bus.listeners[eventType]
	if listeners == nil {
		return
	}

	// Note: This is a simple implementation that removes all instances
	// In production, you might want a more sophisticated approach
	bus.listeners[eventType] = make([]EventListener, 0)
}

// Close gracefully shuts down the event bus and its worker pool
func (bus *InMemoryEventBus) Close() error {
	var closeErr error

	bus.closeOnce.Do(func() {
		logger.Info("🛑 Shutting down event bus worker pool")

		// Signal all workers to stop
		close(bus.closed)

		// Wait for all workers to finish processing current jobs
		done := make(chan struct{})
		go func() {
			bus.workerWg.Wait()
			close(done)
		}()

		// Wait up to 30 seconds for graceful shutdown
		select {
		case <-done:
			logger.Info("✅ Event bus worker pool shut down gracefully")
		case <-time.After(30 * time.Second):
			logger.Warn("⏰ Event bus worker pool shutdown timeout")
			closeErr = errors.New("worker pool shutdown timeout")
		}

		// Close the job queue to prevent new jobs
		close(bus.jobQueue)

		// Drain any remaining jobs
		remaining := 0
		for range bus.jobQueue {
			remaining++
		}

		if remaining > 0 {
			logger.Warn("🗑️ Discarded unprocessed events during shutdown",
				zap.Int("count", remaining))
		}

		logger.Info("🔒 Event bus completely shut down")
	})

	return closeErr
}
