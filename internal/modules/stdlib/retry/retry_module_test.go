package retry

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
// call tests
// =============================================================================

func TestCallSuccessFirstAttempt(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local retry = require("stdlib.retry")
		
		local result, err = retry.call(function()
			return "success"
		end, {max_attempts = 3})
		
		assert(err == nil, "should not error on success: " .. tostring(err))
		assert(result == "success", "should return success")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCallSuccessAfterFailures(t *testing.T) {
	L := newTestState()
	defer L.Close()

	// Use a global table to track attempts since local upvalues don't persist across error()
	err := L.DoString(`
		local retry = require("stdlib.retry")
		_G.attempts = {count = 0}
		
		local result, err = retry.call(function()
			_G.attempts.count = _G.attempts.count + 1
			if _G.attempts.count < 3 then
				error("temporary failure")
			end
			return "success"
		end, {max_attempts = 5})
		
		assert(err == nil, "should succeed after retries: " .. tostring(err))
		assert(result == "success", "should return success")
		assert(_G.attempts.count == 3, "should have tried 3 times, got " .. _G.attempts.count)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCallAllAttemptsFail(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local retry = require("stdlib.retry")
		_G.attempts = {count = 0}
		
		local result, err = retry.call(function()
			_G.attempts.count = _G.attempts.count + 1
			error("always fails")
		end, {max_attempts = 3})
		
		assert(result == nil, "result should be nil when all attempts fail")
		assert(err ~= nil, "should return error when all attempts fail")
		assert(_G.attempts.count == 3, "should have tried max_attempts times, got " .. _G.attempts.count)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCallWithDelay(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local retry = require("stdlib.retry")
		_G.attempts = {count = 0}
		
		local result, err = retry.call(function()
			_G.attempts.count = _G.attempts.count + 1
			if _G.attempts.count < 2 then
				error("temporary failure")
			end
			return "success"
		end, {max_attempts = 3, delay = 10})  -- 10ms delay
		
		assert(err == nil, "should succeed: " .. tostring(err))
		assert(_G.attempts.count == 2, "should have tried twice, got " .. _G.attempts.count)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// exponential tests
// =============================================================================

func TestExponentialBackoff(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local retry = require("stdlib.retry")
		_G.attempts = {count = 0}
		
		local result, err = retry.exponential(function()
			_G.attempts.count = _G.attempts.count + 1
			if _G.attempts.count < 2 then
				error("temporary failure")
			end
			return "success"
		end, {max_attempts = 3, initial_delay = 10})
		
		assert(err == nil, "should succeed: " .. tostring(err))
		assert(_G.attempts.count == 2, "should have tried twice, got " .. _G.attempts.count)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExponentialAllFail(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local retry = require("stdlib.retry")
		_G.attempts = {count = 0}
		
		local result, err = retry.exponential(function()
			_G.attempts.count = _G.attempts.count + 1
			error("always fails")
		end, {max_attempts = 3, initial_delay = 10})
		
		assert(result == nil, "result should be nil")
		assert(err ~= nil, "should return error")
		assert(_G.attempts.count == 3, "should exhaust all attempts, got " .. _G.attempts.count)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// linear tests
// =============================================================================

func TestLinearBackoff(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local retry = require("stdlib.retry")
		_G.attempts = {count = 0}
		
		local result, err = retry.linear(function()
			_G.attempts.count = _G.attempts.count + 1
			if _G.attempts.count < 2 then
				error("temporary failure")
			end
			return "success"
		end, {max_attempts = 3, delay = 10})
		
		assert(err == nil, "should succeed: " .. tostring(err))
		assert(_G.attempts.count == 2, "should have tried twice, got " .. _G.attempts.count)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestLinearAllFail(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local retry = require("stdlib.retry")
		_G.attempts = {count = 0}
		
		local result, err = retry.linear(function()
			_G.attempts.count = _G.attempts.count + 1
			error("always fails")
		end, {max_attempts = 3, delay = 10})
		
		assert(result == nil, "result should be nil")
		assert(err ~= nil, "should return error")
		assert(_G.attempts.count == 3, "should exhaust all attempts, got " .. _G.attempts.count)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// forever tests
// =============================================================================

func TestForeverSuccess(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local retry = require("stdlib.retry")
		_G.attempts = {count = 0}
		
		local result = retry.forever(function()
			_G.attempts.count = _G.attempts.count + 1
			if _G.attempts.count < 3 then
				error("temporary failure")
			end
			return "success"
		end, {delay = 10})
		
		assert(result == "success", "should return success")
		assert(_G.attempts.count == 3, "should have tried 3 times, got " .. _G.attempts.count)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// with_jitter tests
// =============================================================================

func TestWithJitter(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local retry = require("stdlib.retry")
		_G.attempts = {count = 0}
		
		local result, err = retry.with_jitter(function()
			_G.attempts.count = _G.attempts.count + 1
			if _G.attempts.count < 2 then
				error("temporary failure")
			end
			return "success"
		end, {max_attempts = 3, delay = 10, jitter = 0.5})
		
		assert(err == nil, "should succeed: " .. tostring(err))
		assert(_G.attempts.count == 2, "should have tried twice, got " .. _G.attempts.count)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestWithJitterAllFail(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local retry = require("stdlib.retry")
		_G.attempts = {count = 0}
		
		local result, err = retry.with_jitter(function()
			_G.attempts.count = _G.attempts.count + 1
			error("always fails")
		end, {max_attempts = 3, delay = 10, jitter = 0.5})
		
		assert(result == nil, "result should be nil")
		assert(err ~= nil, "should return error")
		assert(_G.attempts.count == 3, "should exhaust all attempts, got " .. _G.attempts.count)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// options tests
// =============================================================================

func TestDelayParsing(t *testing.T) {
	L := newTestState()
	defer L.Close()

	// Test string duration parsing
	err := L.DoString(`
		local retry = require("stdlib.retry")
		
		-- Test with string duration "100ms"
		local result, err = retry.call(function()
			return "success"
		end, {max_attempts = 1, delay = "100ms"})
		
		assert(err == nil, "should succeed with string delay")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDefaultOptions(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local retry = require("stdlib.retry")
		
		-- Test with no options (should use defaults)
		local result, err = retry.call(function()
			return "success"
		end)
		
		assert(err == nil, "should succeed with default options")
		assert(result == "success", "should return success")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
