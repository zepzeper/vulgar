package shell

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
		local shell = require("stdlib.shell")
		local output, err = shell.exec("echo hello")
		assert(err == nil, "exec should not error: " .. tostring(err))
		assert(output ~= nil, "output should not be nil")
		assert(string.find(output, "hello") ~= nil, "should contain hello")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExecWithArgs(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local shell = require("stdlib.shell")
		local output, err = shell.exec("echo hello world")
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
		local shell = require("stdlib.shell")
		local output, err = shell.exec("nonexistent_command_xyz123")
		assert(output == nil or err ~= nil, "should error for invalid command")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// run tests
// =============================================================================

func TestRunSuccess(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local shell = require("stdlib.shell")
		local code, output, err = shell.run("echo test")
		assert(err == nil, "run should not error: " .. tostring(err))
		assert(code == 0, "exit code should be 0")
		assert(string.find(output, "test") ~= nil, "should contain test")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRunFailure(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local shell = require("stdlib.shell")
		local code, output, err = shell.run("exit 1")
		-- exit 1 should return non-zero code
		assert(code ~= 0 or err ~= nil, "should indicate failure")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// pipe tests
// =============================================================================

func TestPipeSimple(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local shell = require("stdlib.shell")
		local output, err = shell.pipe("echo hello", "cat")
		assert(err == nil, "pipe should not error: " .. tostring(err))
		assert(string.find(output, "hello") ~= nil, "should contain hello")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestPipeMultiple(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local shell = require("stdlib.shell")
		local output, err = shell.pipe("echo -e 'a\\nb\\nc'", "sort", "head -1")
		assert(err == nil, "pipe should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// quote tests
// =============================================================================

func TestQuoteSimple(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local shell = require("stdlib.shell")
		local quoted = shell.quote("hello world")
		assert(quoted ~= nil, "quoted should not be nil")
		-- Should be quoted to be safe for shell
		assert(string.find(quoted, "hello") ~= nil, "should contain original text")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQuoteSpecialChars(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local shell = require("stdlib.shell")
		local quoted = shell.quote("test;rm -rf /")
		assert(quoted ~= nil, "quoted should not be nil")
		-- Dangerous chars should be escaped/quoted
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQuoteEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local shell = require("stdlib.shell")
		local quoted = shell.quote("")
		assert(quoted ~= nil, "quoted should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// which tests
// =============================================================================

func TestWhichExists(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local shell = require("stdlib.shell")
		local path, err = shell.which("sh")
		assert(err == nil, "which should not error for sh: " .. tostring(err))
		assert(path ~= nil, "path should not be nil")
		assert(#path > 0, "path should not be empty")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestWhichNotFound(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local shell = require("stdlib.shell")
		local path, err = shell.which("nonexistent_command_xyz123")
		assert(path == nil, "path should be nil for nonexistent command")
		assert(err ~= nil, "should error for nonexistent command")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestWhichCommon(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local shell = require("stdlib.shell")
		
		-- Test common commands
		local commands = {"ls", "cat", "echo"}
		for _, cmd in ipairs(commands) do
			local path, err = shell.which(cmd)
			-- These should exist on most systems
		end
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
