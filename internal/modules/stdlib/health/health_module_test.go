package health

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
// register tests
// =============================================================================

func TestRegister(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local health = require("stdlib.health")
		
		local err = health.register("database", function()
			return {status = "healthy"}
		end)
		
		assert(err == nil, "register should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRegisterMultiple(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local health = require("stdlib.health")
		
		health.register("database", function() return {status = "healthy"} end)
		health.register("cache", function() return {status = "healthy"} end)
		health.register("queue", function() return {status = "healthy"} end)
		
		-- Should register multiple checks
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// check tests
// =============================================================================

func TestCheck(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local health = require("stdlib.health")
		
		health.register("test", function()
			return {status = "healthy", latency = 10}
		end)
		
		local result, err = health.check("test")
		assert(err == nil, "check should not error: " .. tostring(err))
		assert(result ~= nil, "result should not be nil")
		assert(result.status == "healthy", "status should be healthy")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCheckUnregistered(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local health = require("stdlib.health")
		
		local result, err = health.check("nonexistent")
		assert(result == nil or err ~= nil, "should error for unregistered check")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// check_all tests
// =============================================================================

func TestCheckAll(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local health = require("stdlib.health")
		
		health.register("db", function() return {status = "healthy"} end)
		health.register("cache", function() return {status = "healthy"} end)
		
		local results, err = health.check_all()
		assert(err == nil, "check_all should not error: " .. tostring(err))
		assert(results ~= nil, "results should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCheckAllEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local health = require("stdlib.health")
		
		-- Clear any existing checks
		health.clear()
		
		local results, err = health.check_all()
		-- Should return empty or handle gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// status tests
// =============================================================================

func TestStatus(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local health = require("stdlib.health")
		
		health.register("test", function() return {status = "healthy"} end)
		
		local status = health.status()
		assert(status ~= nil, "status should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// unregister tests
// =============================================================================

func TestUnregister(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local health = require("stdlib.health")
		
		health.register("test", function() return {status = "healthy"} end)
		
		local err = health.unregister("test")
		assert(err == nil, "unregister should not error: " .. tostring(err))
		
		local result, checkErr = health.check("test")
		assert(result == nil or checkErr ~= nil, "check should fail after unregister")
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
		local health = require("stdlib.health")
		
		health.register("test1", function() return {status = "healthy"} end)
		health.register("test2", function() return {status = "healthy"} end)
		
		health.clear()
		
		-- All checks should be cleared
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// is_healthy tests
// =============================================================================

func TestIsHealthy(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local health = require("stdlib.health")
		
		health.register("test", function() return {status = "healthy"} end)
		
		local healthy = health.is_healthy()
		assert(healthy == true, "should be healthy")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIsHealthyWithUnhealthy(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local health = require("stdlib.health")
		
		health.clear()
		health.register("healthy", function() return {status = "healthy"} end)
		health.register("unhealthy", function() return {status = "unhealthy"} end)
		
		local healthy = health.is_healthy()
		assert(healthy == false, "should not be healthy when one check fails")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
