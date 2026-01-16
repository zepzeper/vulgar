package util

import (
	"sync/atomic"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// Registry key for storing the EventQueue in the Lua state
const EventQueueRegistryKey = "vulgar_event_queue"

// GetEventQueue retrieves the EventQueue from the Lua state
func GetEventQueue(L *lua.LState) *EventQueue {
	// L.Get(lua.RegistryIndex) returns the registry table
	registry := L.Get(lua.RegistryIndex)
	if tbl, ok := registry.(*lua.LTable); ok {
		ud := L.GetField(tbl, EventQueueRegistryKey)
		if v, ok := ud.(*lua.LUserData); ok {
			if q, ok := v.Value.(*EventQueue); ok {
				return q
			}
		}
	}
	return nil
}

// EventData represents a single event to be processed
type EventData struct {
	Callback *lua.LFunction
	Data     lua.LValue
	// Task is a function to execute on the main thread.
	// If Task is provided, Callback and Data are ignored.
	Task func(*lua.LState)
}

// EventQueue provides a thread-safe way to queue Lua callbacks from goroutines
// and process them on the main Lua thread.
//
// Usage:
//
//		queue := util.NewEventQueue(L)
//
//		// From goroutine (thread-safe):
//		queue.AddSource() // Increment active source count
//		queue.Queue(callback, eventData)
//	 // OR
//	 queue.QueueTask(func(L *lua.LState) { ... })
//
//		// From main Lua thread:
//		queue.WaitForEvents() // Block until events arrive or timeout
//
// Thread Safety:
//   - Queue(), AddSource(), RemoveSource() can be called from any goroutine
//   - WaitForEvents(), Process() MUST be called from the main Lua thread only
//   - Close() can be called from any goroutine
type EventQueue struct {
	events        chan EventData
	L             *lua.LState
	done          chan struct{}
	activeSources int32
}

// NewEventQueue creates a new event queue for the given Lua state
// bufferSize controls how many events can be queued before dropping (default: 100)
func NewEventQueue(L *lua.LState, bufferSize int) *EventQueue {
	if bufferSize <= 0 {
		bufferSize = 100 // Default buffer size
	}
	return &EventQueue{
		events: make(chan EventData, bufferSize),
		L:      L,
		done:   make(chan struct{}),
	}
}

// AddSource increments the count of active async sources (e.g. running timers)
func (q *EventQueue) AddSource() {
	atomic.AddInt32(&q.activeSources, 1)
}

// RemoveSource decrements the count of active async sources
func (q *EventQueue) RemoveSource() {
	atomic.AddInt32(&q.activeSources, -1)
}

// HasActiveSources returns true if there are any active async sources
func (q *EventQueue) HasActiveSources() bool {
	return atomic.LoadInt32(&q.activeSources) > 0
}

// Queue safely adds an event to the queue (can be called from any goroutine)
// If the queue is full, the event is dropped (non-blocking)
// If the queue is closed, the event is ignored
func (q *EventQueue) Queue(callback *lua.LFunction, data lua.LValue) {
	q.send(EventData{Callback: callback, Data: data})
}

// QueueTask safely adds a generic task to be executed on the main thread
func (q *EventQueue) QueueTask(task func(*lua.LState)) {
	q.send(EventData{Task: task})
}

func (q *EventQueue) send(event EventData) {
	select {
	case q.events <- event:
		// Successfully queued
	case <-q.done:
		// Queue closed, ignore event
	default:
		// Channel full - drop event (non-blocking)
		// In production, you might want to log this
	}
}

// WaitForEvents blocks until an event is available or the queue is closed.
// It returns true if an event was processed, false if the queue is closed or no sources remain.
func (q *EventQueue) WaitForEvents() bool {
	// If no active sources and no pending events, we're done
	if !q.HasActiveSources() && len(q.events) == 0 {
		return false
	}

	select {
	case event := <-q.events:
		q.processEvent(event)
		return true
	case <-q.done:
		return false
	// Add a small timeout to allow checking for active sources periodically
	// This prevents hanging if a source is removed without firing an event
	case <-time.After(100 * time.Millisecond):
		return q.HasActiveSources()
	}
}

// processEvent executes a single event callback
func (q *EventQueue) processEvent(event EventData) {
	// Safe to call Lua here - we're in main thread

	if event.Task != nil {
		event.Task(q.L)
		return
	}

	q.L.Push(event.Callback)
	if event.Data != nil {
		q.L.Push(event.Data)
		if err := q.L.PCall(1, 0, nil); err != nil {
			// Error in callback - log but don't crash
			_ = err
		}
	} else {
		if err := q.L.PCall(0, 0, nil); err != nil {
			_ = err
		}
	}
}

// Process handles queued events (MUST be called from main Lua thread only)
// Processes all available events and returns immediately if none are available (non-blocking)
// Returns the number of events processed
func (q *EventQueue) Process() int {
	count := 0
	for {
		select {
		case event := <-q.events:
			q.processEvent(event)
			count++
		case <-q.done:
			// Queue closed, stop processing
			return count
		default:
			// No more events available, return immediately (non-blocking)
			return count
		}
	}
}

// Close closes the event queue and prevents new events from being queued
// Can be called from any goroutine
func (q *EventQueue) Close() {
	select {
	case <-q.done:
		// Already closed
	default:
		close(q.done)
	}
}
