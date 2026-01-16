package path

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func setupLuaState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

func TestJoinTwoPaths(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.join("home", "user")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := filepath.Join("home", "user")
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestJoinMultiplePaths(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.join("home", "user", "docs", "file.txt")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := filepath.Join("home", "user", "docs", "file.txt")
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestJoinSinglePath(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.join("only")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != "only" {
		t.Errorf("expected 'only', got '%s'", result)
	}
}

func TestDir(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.dir("/home/user/file.txt")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := filepath.Dir("/home/user/file.txt")
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestDirNoDir(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.dir("file.txt")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := filepath.Dir("file.txt")
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestBase(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.base("/home/user/file.txt")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != "file.txt" {
		t.Errorf("expected 'file.txt', got '%s'", result)
	}
}

func TestBaseNoPath(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.base("file.txt")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != "file.txt" {
		t.Errorf("expected 'file.txt', got '%s'", result)
	}
}

func TestExt(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.ext("file.txt")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != ".txt" {
		t.Errorf("expected '.txt', got '%s'", result)
	}
}

func TestExtNoExt(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.ext("filename")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestExtMultipleDots(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.ext("archive.tar.gz")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != ".gz" {
		t.Errorf("expected '.gz', got '%s'", result)
	}
}

func TestAbs(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result, err = path.abs(".")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	errVal := L.GetGlobal("err")

	if errVal != lua.LNil {
		t.Errorf("unexpected error: %v", errVal)
	}

	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got '%s'", result)
	}
}

func TestClean(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.clean("/home/user/../user/./docs")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := filepath.Clean("/home/user/../user/./docs")
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestCleanDots(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.clean("./a/../b/./c")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := filepath.Clean("./a/../b/./c")
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestSplit(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		dir, file = path.split("/home/user/file.txt")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	dir := L.GetGlobal("dir").String()
	file := L.GetGlobal("file").String()

	expectedDir, expectedFile := filepath.Split("/home/user/file.txt")
	if dir != expectedDir {
		t.Errorf("expected dir '%s', got '%s'", expectedDir, dir)
	}
	if file != expectedFile {
		t.Errorf("expected file '%s', got '%s'", expectedFile, file)
	}
}

func TestSplitNoDir(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		dir, file = path.split("file.txt")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	dir := L.GetGlobal("dir").String()
	file := L.GetGlobal("file").String()

	if dir != "" {
		t.Errorf("expected empty dir, got '%s'", dir)
	}
	if file != "file.txt" {
		t.Errorf("expected 'file.txt', got '%s'", file)
	}
}

func TestIsAbsTrue(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	var testPath string
	if runtime.GOOS == "windows" {
		testPath = "C:\\\\Users"
	} else {
		testPath = "/home/user"
	}

	L.SetGlobal("test_path", lua.LString(testPath))
	err := L.DoString(`
		local path = require("path")
		result = path.is_abs(test_path)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result != lua.LTrue {
		t.Error("expected true for absolute path")
	}
}

func TestIsAbsFalse(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.is_abs("relative/path")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result != lua.LFalse {
		t.Error("expected false for relative path")
	}
}

func TestSeparator(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.separator
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	expected := string(filepath.Separator)
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestJoinWithSeparator(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local path = require("path")
		result = path.join("a", "b")
		has_sep = string.find(result, path.separator, 1, true) ~= nil
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	hasSep := L.GetGlobal("has_sep")

	if !strings.Contains(result, string(filepath.Separator)) {
		t.Errorf("joined path should contain separator: %s", result)
	}
	if hasSep != lua.LTrue {
		t.Error("Lua should detect separator in joined path")
	}
}

func TestLoaderReturnsModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`path = require("path")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	mod := L.GetGlobal("path")
	if mod.Type() != lua.LTTable {
		t.Errorf("expected table, got %s", mod.Type())
	}
}

func TestLoaderExportsAllFunctions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`path = require("path")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	tbl := L.GetGlobal("path").(*lua.LTable)

	funcs := []string{"join", "dir", "base", "ext", "abs", "clean", "split", "is_abs"}
	for _, name := range funcs {
		fn := L.GetField(tbl, name)
		if fn.Type() != lua.LTFunction {
			t.Errorf("expected %s to be a function", name)
		}
	}

	// Check separator constant
	sep := L.GetField(tbl, "separator")
	if sep.Type() != lua.LTString {
		t.Errorf("expected separator to be a string, got %s", sep.Type())
	}
}
