package fs

import (
	"os"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func setupLuaState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

func TestReadFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_read_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "test content"
	tmpFile.WriteString(content)
	tmpFile.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_path", lua.LString(tmpFile.Name()))
	luaErr := L.DoString(`
		local fs = require("fs")
		result, err = fs.read_file(test_path)
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	result := L.GetGlobal("result").String()
	errVal := L.GetGlobal("err")

	if result != content {
		t.Errorf("expected '%s', got '%s'", content, result)
	}
	if errVal != lua.LNil {
		t.Errorf("unexpected error: %v", errVal)
	}
}

func TestReadFileNotFound(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	luaErr := L.DoString(`
		local fs = require("fs")
		result, err = fs.read_file("/nonexistent/path/to/file.txt")
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	result := L.GetGlobal("result")
	errVal := L.GetGlobal("err")

	if result != lua.LNil {
		t.Error("expected nil result for nonexistent file")
	}
	if errVal == lua.LNil {
		t.Error("expected error for nonexistent file")
	}
}

func TestWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test_write.txt")

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_path", lua.LString(testPath))
	luaErr := L.DoString(`
		local fs = require("fs")
		err = fs.write_file(test_path, "written content")
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	content, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if string(content) != "written content" {
		t.Errorf("expected 'written content', got '%s'", content)
	}
}

func TestAppendFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_append_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("initial")
	tmpFile.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_path", lua.LString(tmpFile.Name()))
	luaErr := L.DoString(`
		local fs = require("fs")
		err = fs.append_file(test_path, " appended")
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	content, _ := os.ReadFile(tmpFile.Name())
	if string(content) != "initial appended" {
		t.Errorf("expected 'initial appended', got '%s'", content)
	}
}

func TestExistsWithExistingFile(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test_exists_*.txt")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_path", lua.LString(tmpFile.Name()))
	luaErr := L.DoString(`
		local fs = require("fs")
		result = fs.exists(test_path)
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	result := L.GetGlobal("result")
	if result != lua.LTrue {
		t.Error("expected true for existing file")
	}
}

func TestExistsWithNonexistentFile(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	luaErr := L.DoString(`
		local fs = require("fs")
		result = fs.exists("/nonexistent/path/to/file.txt")
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	result := L.GetGlobal("result")
	if result != lua.LFalse {
		t.Error("expected false for nonexistent file")
	}
}

func TestRemoveFile(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test_remove_*.txt")
	tmpFile.Close()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_path", lua.LString(tmpFile.Name()))
	luaErr := L.DoString(`
		local fs = require("fs")
		err = fs.remove(test_path)
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	if _, err := os.Stat(tmpFile.Name()); !os.IsNotExist(err) {
		t.Error("expected file to be removed")
	}
}

func TestMkdir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "new_dir", "nested")

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_path", lua.LString(newDir))
	luaErr := L.DoString(`
		local fs = require("fs")
		err = fs.mkdir(test_path)
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}
}

func TestListDir(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("2"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_path", lua.LString(tmpDir))
	luaErr := L.DoString(`
		local fs = require("fs")
		result, err = fs.list_dir(test_path)
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	result := L.GetGlobal("result").(*lua.LTable)
	if result.Len() != 3 {
		t.Errorf("expected 3 entries, got %d", result.Len())
	}
}

func TestListDirReturnsEntryInfo(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644)

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_path", lua.LString(tmpDir))
	luaErr := L.DoString(`
		local fs = require("fs")
		result, err = fs.list_dir(test_path)
		entry = result[1]
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	entry := L.GetGlobal("entry").(*lua.LTable)
	name := L.GetField(entry, "name")
	isDir := L.GetField(entry, "is_dir")

	if name.String() != "file.txt" {
		t.Errorf("expected name 'file.txt', got '%s'", name.String())
	}
	if isDir != lua.LFalse {
		t.Error("expected is_dir false for file")
	}
}

func TestCopy(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	os.WriteFile(srcPath, []byte("copy me"), 0644)

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("src", lua.LString(srcPath))
	L.SetGlobal("dst", lua.LString(dstPath))
	luaErr := L.DoString(`
		local fs = require("fs")
		err = fs.copy(src, dst)
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	content, _ := os.ReadFile(dstPath)
	if string(content) != "copy me" {
		t.Errorf("expected 'copy me', got '%s'", content)
	}
}

func TestMove(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "moved.txt")

	os.WriteFile(srcPath, []byte("move me"), 0644)

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("src", lua.LString(srcPath))
	L.SetGlobal("dst", lua.LString(dstPath))
	luaErr := L.DoString(`
		local fs = require("fs")
		err = fs.move(src, dst)
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Error("expected source to be removed")
	}

	content, _ := os.ReadFile(dstPath)
	if string(content) != "move me" {
		t.Errorf("expected 'move me', got '%s'", content)
	}
}

func TestStat(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test_stat_*.txt")
	tmpFile.WriteString("stat content")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("test_path", lua.LString(tmpFile.Name()))
	luaErr := L.DoString(`
		local fs = require("fs")
		result, err = fs.stat(test_path)
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	result := L.GetGlobal("result").(*lua.LTable)
	size := L.GetField(result, "size").(lua.LNumber)
	isDir := L.GetField(result, "is_dir")

	if size != 12 {
		t.Errorf("expected size 12, got %v", size)
	}
	if isDir != lua.LFalse {
		t.Error("expected is_dir false")
	}
}

func TestStatNonexistent(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	luaErr := L.DoString(`
		local fs = require("fs")
		result, err = fs.stat("/nonexistent/path")
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	result := L.GetGlobal("result")
	errVal := L.GetGlobal("err")

	if result != lua.LNil {
		t.Error("expected nil result")
	}
	if errVal == lua.LNil {
		t.Error("expected error")
	}
}

func TestHandleNew(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	luaErr := L.DoString(`
		local fs = require("fs")
		handle = fs.new({ base_dir = "/tmp" })
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	handle := L.GetGlobal("handle")
	if handle.Type() != lua.LTUserData {
		t.Errorf("expected userdata, got %s", handle.Type())
	}
}

func TestHandleWithBaseDir(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("base dir content"), 0644)

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("base_dir", lua.LString(tmpDir))
	luaErr := L.DoString(`
		local fs = require("fs")
		local handle = fs.new({ base_dir = base_dir })
		result, err = handle:read_file("test.txt")
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	result := L.GetGlobal("result").String()
	if result != "base dir content" {
		t.Errorf("expected 'base dir content', got '%s'", result)
	}
}

func TestHandleWriteFile(t *testing.T) {
	tmpDir := t.TempDir()

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("base_dir", lua.LString(tmpDir))
	luaErr := L.DoString(`
		local fs = require("fs")
		local handle = fs.new({ base_dir = base_dir })
		err = handle:write_file("output.txt", "handle write")
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "output.txt"))
	if string(content) != "handle write" {
		t.Errorf("expected 'handle write', got '%s'", content)
	}
}

func TestHandleExists(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "exists.txt"), []byte(""), 0644)

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("base_dir", lua.LString(tmpDir))
	luaErr := L.DoString(`
		local fs = require("fs")
		local handle = fs.new({ base_dir = base_dir })
		result = handle:exists("exists.txt")
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	result := L.GetGlobal("result")
	if result != lua.LTrue {
		t.Error("expected true")
	}
}

func TestHandleListDir(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte(""), 0644)

	L := setupLuaState()
	defer L.Close()

	L.SetGlobal("base_dir", lua.LString(tmpDir))
	luaErr := L.DoString(`
		local fs = require("fs")
		local handle = fs.new({ base_dir = base_dir })
		result, err = handle:list_dir(".")
	`)
	if luaErr != nil {
		t.Fatalf("failed to execute: %v", luaErr)
	}

	result := L.GetGlobal("result").(*lua.LTable)
	if result.Len() != 2 {
		t.Errorf("expected 2 entries, got %d", result.Len())
	}
}

func TestLoaderReturnsModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`fs = require("fs")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	mod := L.GetGlobal("fs")
	if mod.Type() != lua.LTTable {
		t.Errorf("expected table, got %s", mod.Type())
	}
}

func TestLoaderExportsAllFunctions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`fs = require("fs")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	tbl := L.GetGlobal("fs").(*lua.LTable)

	funcs := []string{"new", "read_file", "write_file", "append_file", "exists", "remove", "mkdir", "list_dir", "copy", "move", "stat"}
	for _, name := range funcs {
		fn := L.GetField(tbl, name)
		if fn.Type() != lua.LTFunction {
			t.Errorf("expected %s to be a function", name)
		}
	}
}
