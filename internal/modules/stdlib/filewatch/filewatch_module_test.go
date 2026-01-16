package filewatch

import (
	"os"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

// =============================================================================
// watch tests
// =============================================================================

func TestWatchFile(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("initial"), 0644)

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local filewatch = require("stdlib.filewatch")
		local changes = 0
		
		local watcher, err = filewatch.watch(test_file, function(event)
			changes = changes + 1
		end)
		
		assert(err == nil, "watch should not error: " .. tostring(err))
		assert(watcher ~= nil, "watcher should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestWatchDirectory(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	L.SetGlobal("test_dir", lua.LString(tmpDir))

	err := L.DoString(`
		local filewatch = require("stdlib.filewatch")
		
		local watcher, err = filewatch.watch(test_dir, function(event)
			-- Handle directory changes
		end)
		
		assert(err == nil, "watch directory should not error: " .. tostring(err))
		assert(watcher ~= nil, "watcher should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestWatchNonexistent(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local filewatch = require("stdlib.filewatch")
		
		local watcher, err = filewatch.watch("/nonexistent/path/xyz", function() end)
		assert(watcher == nil, "watcher should be nil for nonexistent path")
		assert(err ~= nil, "should error for nonexistent path")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// unwatch tests
// =============================================================================

func TestUnwatch(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("content"), 0644)

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local filewatch = require("stdlib.filewatch")
		
		local watcher, _ = filewatch.watch(test_file, function() end)
		local err = filewatch.unwatch(watcher)
		
		assert(err == nil, "unwatch should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestUnwatchNil(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local filewatch = require("stdlib.filewatch")
		local err = filewatch.unwatch(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// watch_glob tests
// =============================================================================

func TestWatchGlob(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("2"), 0644)

	L.SetGlobal("test_pattern", lua.LString(filepath.Join(tmpDir, "*.txt")))

	err := L.DoString(`
		local filewatch = require("stdlib.filewatch")
		
		local watcher, err = filewatch.watch_glob(test_pattern, function(event)
			-- Handle file changes
		end)
		
		assert(err == nil, "watch_glob should not error: " .. tostring(err))
		assert(watcher ~= nil, "watcher should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestWatchGlobNoMatches(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	L.SetGlobal("test_pattern", lua.LString(filepath.Join(tmpDir, "*.nonexistent")))

	err := L.DoString(`
		local filewatch = require("stdlib.filewatch")
		
		local watcher, err = filewatch.watch_glob(test_pattern, function() end)
		-- Should handle no matches - either nil watcher or error
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestWatchGlobInvalidPattern(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local filewatch = require("stdlib.filewatch")
		
		local watcher, err = filewatch.watch_glob("[invalid", function() end)
		-- Invalid glob should error
		assert(watcher == nil or err ~= nil, "should handle invalid glob")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
