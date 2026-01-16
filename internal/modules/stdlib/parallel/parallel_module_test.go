package parallel

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
// map tests
// =============================================================================

func TestMapSimple(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local parallel = require("stdlib.parallel")
		local items = {1, 2, 3, 4, 5}
		
		local results, err = parallel.map(items, function(item)
			return item * 2
		end)
		
		assert(err == nil, "map should not error: " .. tostring(err))
		assert(#results == 5, "should have 5 results")
		assert(results[1] == 2, "first result should be 2")
		assert(results[5] == 10, "last result should be 10")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMapEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local parallel = require("stdlib.parallel")
		local results, err = parallel.map({}, function(item)
			return item * 2
		end)
		
		assert(err == nil, "map empty should not error")
		assert(#results == 0, "should have 0 results")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMapWithError(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local parallel = require("stdlib.parallel")
		local items = {1, 2, 3}
		
		local results, err = parallel.map(items, function(item)
			if item == 2 then
				error("failed on item 2")
			end
			return item
		end)
		
		-- Should return error when any item fails
		assert(err ~= nil or results == nil, "should handle error in worker")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// each tests
// =============================================================================

func TestEachSimple(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local parallel = require("stdlib.parallel")
		local items = {1, 2, 3}
		local processed = 0
		
		local err = parallel.each(items, function(item)
			processed = processed + 1
		end)
		
		assert(err == nil, "each should not error: " .. tostring(err))
		assert(processed == 3, "should process all items")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEachEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local parallel = require("stdlib.parallel")
		local err = parallel.each({}, function(item) end)
		assert(err == nil, "each empty should not error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// all tests
// =============================================================================

func TestAllSuccess(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local parallel = require("stdlib.parallel")
		
		local results, err = parallel.all({
			function() return "a" end,
			function() return "b" end,
			function() return "c" end
		})
		
		assert(err == nil, "all should not error: " .. tostring(err))
		assert(#results == 3, "should have 3 results")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestAllWithFailure(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local parallel = require("stdlib.parallel")
		
		local results, err = parallel.all({
			function() return "ok" end,
			function() error("failed") end,
			function() return "ok" end
		})
		
		-- Should return error when any function fails
		assert(err ~= nil or results == nil, "should handle failure")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestAllEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local parallel = require("stdlib.parallel")
		local results, err = parallel.all({})
		assert(err == nil, "all empty should not error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// any tests
// =============================================================================

func TestAnyFirstWins(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local parallel = require("stdlib.parallel")
		
		local result, err = parallel.any({
			function() return "first" end,
			function() return "second" end
		})
		
		assert(err == nil, "any should not error: " .. tostring(err))
		assert(result == "first" or result == "second", "should return one result")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestAnyAllFail(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local parallel = require("stdlib.parallel")
		
		local result, err = parallel.any({
			function() error("fail1") end,
			function() error("fail2") end
		})
		
		-- When all fail, should return error
		assert(err ~= nil, "should error when all fail")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// pool tests
// =============================================================================

func TestPoolCreate(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local parallel = require("stdlib.parallel")
		
		local pool, err = parallel.pool(4)
		assert(err == nil, "pool should not error: " .. tostring(err))
		assert(pool ~= nil, "pool should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestPoolZeroWorkers(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local parallel = require("stdlib.parallel")
		
		local pool, err = parallel.pool(0)
		-- Zero workers should either error or default to 1
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
