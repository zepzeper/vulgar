package process

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
// exec tests
// =============================================================================

func TestExecSimple(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local process = require("stdlib.process")
		local output, err = process.exec("echo", {"hello"})
		assert(err == nil, "exec should not error: " .. tostring(err))
		assert(output ~= nil, "output should not be nil")
		assert(string.find(output, "hello") ~= nil, "should contain hello")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExecWithMultipleArgs(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local process = require("stdlib.process")
		local output, err = process.exec("echo", {"hello", "world"})
		assert(err == nil, "exec should not error")
		assert(string.find(output, "hello") ~= nil, "should contain hello")
		assert(string.find(output, "world") ~= nil, "should contain world")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExecInvalidCommand(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local process = require("stdlib.process")
		local output, err = process.exec("nonexistent_command_xyz", {})
		assert(output == nil or err ~= nil, "should error for invalid command")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExecNoArgs(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local process = require("stdlib.process")
		local output, err = process.exec("pwd", {})
		assert(err == nil, "exec pwd should not error: " .. tostring(err))
		assert(output ~= nil, "output should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// spawn tests
// =============================================================================

func TestSpawnProcess(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local process = require("stdlib.process")
		local proc, err = process.spawn("sleep", {"0.1"})
		assert(err == nil, "spawn should not error: " .. tostring(err))
		assert(proc ~= nil, "proc should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSpawnInvalidCommand(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local process = require("stdlib.process")
		local proc, err = process.spawn("nonexistent_xyz", {})
		assert(proc == nil or err ~= nil, "should error for invalid command")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// kill tests
// =============================================================================

func TestKillInvalidPid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local process = require("stdlib.process")
		local err = process.kill(999999999)
		-- Should error for invalid/nonexistent PID
		assert(err ~= nil, "should error for invalid PID")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list tests
// =============================================================================

func TestListProcesses(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local process = require("stdlib.process")
		local procs, err = process.list()
		assert(err == nil, "list should not error: " .. tostring(err))
		assert(procs ~= nil, "procs should not be nil")
		assert(#procs > 0, "should have at least 1 process")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// pid tests
// =============================================================================

func TestPid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local process = require("stdlib.process")
		local pid = process.pid()
		assert(pid ~= nil, "pid should not be nil")
		assert(pid > 0, "pid should be positive")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// env tests
// =============================================================================

func TestEnvGet(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local process = require("stdlib.process")
		local path = process.env("PATH")
		assert(path ~= nil, "PATH should exist")
		assert(#path > 0, "PATH should not be empty")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEnvGetMissing(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local process = require("stdlib.process")
		local value = process.env("NONEXISTENT_VAR_XYZ123")
		assert(value == nil or value == "", "missing env var should be nil or empty")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEnvSet(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local process = require("stdlib.process")
		process.env("TEST_VAR", "test_value")
		local value = process.env("TEST_VAR")
		assert(value == "test_value", "should retrieve set value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
