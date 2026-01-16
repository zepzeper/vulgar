package queue

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
// new tests
// =============================================================================

func TestNew(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local queue = require("stdlib.queue")
		local q, err = queue.new()
		assert(err == nil, "new should not error: " .. tostring(err))
		assert(q ~= nil, "queue should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestNewWithOptions(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local queue = require("stdlib.queue")
		local q, err = queue.new({
			max_size = 100,
			name = "test_queue"
		})
		assert(err == nil, "new with options should not error: " .. tostring(err))
		assert(q ~= nil, "queue should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// push tests
// =============================================================================

func TestPush(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local queue = require("stdlib.queue")
		local q, _ = queue.new()
		
		local err = queue.push(q, "item1")
		assert(err == nil, "push should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestPushMultiple(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local queue = require("stdlib.queue")
		local q, _ = queue.new()
		
		queue.push(q, "item1")
		queue.push(q, "item2")
		queue.push(q, "item3")
		
		assert(queue.size(q) == 3, "should have 3 items")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// pop tests
// =============================================================================

func TestPop(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local queue = require("stdlib.queue")
		local q, _ = queue.new()
		
		queue.push(q, "item1")
		queue.push(q, "item2")
		
		local item, err = queue.pop(q)
		assert(err == nil, "pop should not error: " .. tostring(err))
		assert(item == "item1", "should get first item")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestPopEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local queue = require("stdlib.queue")
		local q, _ = queue.new()
		
		local item, err = queue.pop(q)
		assert(item == nil, "item should be nil for empty queue")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// peek tests
// =============================================================================

func TestPeek(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local queue = require("stdlib.queue")
		local q, _ = queue.new()
		
		queue.push(q, "item1")
		
		local item = queue.peek(q)
		assert(item == "item1", "should peek first item")
		assert(queue.size(q) == 1, "size should still be 1")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestPeekEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local queue = require("stdlib.queue")
		local q, _ = queue.new()
		
		local item = queue.peek(q)
		assert(item == nil, "peek should return nil for empty queue")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// size tests
// =============================================================================

func TestSize(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local queue = require("stdlib.queue")
		local q, _ = queue.new()
		
		assert(queue.size(q) == 0, "empty queue should have size 0")
		
		queue.push(q, "item1")
		assert(queue.size(q) == 1, "should have size 1")
		
		queue.push(q, "item2")
		assert(queue.size(q) == 2, "should have size 2")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// is_empty tests
// =============================================================================

func TestIsEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local queue = require("stdlib.queue")
		local q, _ = queue.new()
		
		assert(queue.is_empty(q) == true, "new queue should be empty")
		
		queue.push(q, "item1")
		assert(queue.is_empty(q) == false, "queue with item should not be empty")
		
		queue.pop(q)
		assert(queue.is_empty(q) == true, "queue after pop should be empty")
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
		local queue = require("stdlib.queue")
		local q, _ = queue.new()
		
		queue.push(q, "item1")
		queue.push(q, "item2")
		
		queue.clear(q)
		
		assert(queue.size(q) == 0, "cleared queue should have size 0")
		assert(queue.is_empty(q) == true, "cleared queue should be empty")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// to_array tests
// =============================================================================

func TestToArray(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local queue = require("stdlib.queue")
		local q, _ = queue.new()
		
		queue.push(q, "item1")
		queue.push(q, "item2")
		queue.push(q, "item3")
		
		local arr = queue.to_array(q)
		assert(#arr == 3, "array should have 3 items")
		assert(arr[1] == "item1", "first item should match")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
