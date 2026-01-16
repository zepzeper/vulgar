package timer

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
// after tests
// =============================================================================

func TestAfter(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local timer = require("stdlib.timer")
		local executed = false
		
		local t, err = timer.after(10, function()
			executed = true
		end)
		
		assert(err == nil, "after should not error: " .. tostring(err))
		assert(t ~= nil, "timer should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestAfterZero(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local timer = require("stdlib.timer")
		
		local t, err = timer.after(0, function() end)
		-- Zero delay should either work immediately or error
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestAfterNegative(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local timer = require("stdlib.timer")
		
		local t, err = timer.after(-100, function() end)
		-- Negative delay should error or be handled
		assert(t == nil or err ~= nil, "should handle negative delay")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// every tests
// =============================================================================

func TestEvery(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local timer = require("stdlib.timer")
		local count = 0
		
		local t, err = timer.every(10, function()
			count = count + 1
		end)
		
		assert(err == nil, "every should not error: " .. tostring(err))
		assert(t ~= nil, "timer should not be nil")
		
		-- Stop it immediately
		timer.stop(t)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEveryNegative(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local timer = require("stdlib.timer")
		
		local t, err = timer.every(-10, function() end)
		assert(t == nil or err ~= nil, "should handle negative interval")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// stop tests
// =============================================================================

func TestStopNil(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local timer = require("stdlib.timer")
		local err = timer.stop(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestStopValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local timer = require("stdlib.timer")
		
		local t, _ = timer.after(1000, function() end)
		local err = timer.stop(t)
		
		assert(err == nil, "stop should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// sleep tests
// =============================================================================

func TestSleep(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local timer = require("stdlib.timer")
		
		timer.sleep(1) -- Sleep for 1ms
		-- Should complete without error
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSleepZero(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local timer = require("stdlib.timer")
		timer.sleep(0)
		-- Zero sleep should work
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// now tests
// =============================================================================

func TestNow(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local timer = require("stdlib.timer")
		
		local timestamp = timer.now()
		assert(timestamp ~= nil, "timestamp should not be nil")
		assert(timestamp > 0, "timestamp should be positive")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// measure tests
// =============================================================================

func TestMeasure(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local timer = require("stdlib.timer")
		
		local duration, result = timer.measure(function()
			return "test"
		end)
		
		assert(duration ~= nil, "duration should not be nil")
		assert(duration >= 0, "duration should be non-negative")
		assert(result == "test", "result should be test")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMeasureWithError(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local timer = require("stdlib.timer")
		
		local duration, result, err = timer.measure(function()
			error("test error")
		end)
		
		-- Should capture error
		assert(err ~= nil, "should capture error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
