package event

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

// =============================================================================
// on tests
// =============================================================================

func TestOn(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local event = require("stdlib.event")
		local received = nil
		
		local unsub = event.on("test", function(data)
			received = data
		end)
		
		assert(unsub ~= nil, "unsub should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestOnMultiple(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local event = require("stdlib.event")
		local count = 0
		
		event.on("test", function() count = count + 1 end)
		event.on("test", function() count = count + 1 end)
		
		-- Should register multiple handlers
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// emit tests
// =============================================================================

func TestEmit(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local event = require("stdlib.event")
		local received = nil
		
		event.on("test", function(data)
			received = data
		end)
		
		event.emit("test", "hello")
		assert(received == "hello", "should receive emitted data")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEmitNoHandlers(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local event = require("stdlib.event")
		
		-- Emit with no handlers should not error
		event.emit("unknown_event", "data")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEmitWithTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local event = require("stdlib.event")
		local received = nil
		
		event.on("test", function(data)
			received = data
		end)
		
		event.emit("test", {name = "John", age = 30})
		assert(received ~= nil, "should receive data")
		assert(received.name == "John", "name should match")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// once tests
// =============================================================================

func TestOnce(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local event = require("stdlib.event")
		local count = 0
		
		event.once("test", function()
			count = count + 1
		end)
		
		event.emit("test")
		event.emit("test")
		
		assert(count == 1, "should only fire once")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// off tests
// =============================================================================

func TestOff(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local event = require("stdlib.event")
		local count = 0
		
		local handler = function()
			count = count + 1
		end
		
		event.on("test", handler)
		event.emit("test")
		
		event.off("test", handler)
		event.emit("test")
		
		assert(count == 1, "handler should be removed")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestOffAll(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local event = require("stdlib.event")
		local count = 0
		
		event.on("test", function() count = count + 1 end)
		event.on("test", function() count = count + 1 end)
		
		event.off("test") -- Remove all handlers for event
		event.emit("test")
		
		assert(count == 0, "all handlers should be removed")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// clear tests
// =============================================================================

func TestClear(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local event = require("stdlib.event")
		local count = 0
		
		event.on("test1", function() count = count + 1 end)
		event.on("test2", function() count = count + 1 end)
		
		event.clear() -- Clear all events
		
		event.emit("test1")
		event.emit("test2")
		
		assert(count == 0, "all events should be cleared")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// listeners tests
// =============================================================================

func TestListeners(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local event = require("stdlib.event")
		
		event.on("test", function() end)
		event.on("test", function() end)
		
		local count = event.listeners("test")
		assert(count == 2, "should have 2 listeners")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestListenersEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local event = require("stdlib.event")
		
		local count = event.listeners("nonexistent")
		assert(count == 0, "should have 0 listeners")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
