package tar

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
// create tests
// =============================================================================

func TestCreateFromFile(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	tarFile := filepath.Join(tmpDir, "archive.tar")

	if err := os.WriteFile(srcFile, []byte("Hello, World!"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	L.SetGlobal("src_file", lua.LString(srcFile))
	L.SetGlobal("tar_file", lua.LString(tarFile))

	err := L.DoString(`
		local tar = require("stdlib.tar")
		local err = tar.create(tar_file, {src_file})
		assert(err == nil, "create should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}

	if _, err := os.Stat(tarFile); os.IsNotExist(err) {
		t.Fatal("tar file was not created")
	}
}

func TestCreateFromMultipleFiles(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	tarFile := filepath.Join(tmpDir, "archive.tar")

	os.WriteFile(file1, []byte("Content 1"), 0644)
	os.WriteFile(file2, []byte("Content 2"), 0644)

	L.SetGlobal("file1", lua.LString(file1))
	L.SetGlobal("file2", lua.LString(file2))
	L.SetGlobal("tar_file", lua.LString(tarFile))

	err := L.DoString(`
		local tar = require("stdlib.tar")
		local err = tar.create(tar_file, {file1, file2})
		assert(err == nil, "create should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCreateFromDirectory(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "source")
	os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("Content"), 0644)
	os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("Nested"), 0644)

	tarFile := filepath.Join(tmpDir, "archive.tar")

	L.SetGlobal("src_dir", lua.LString(srcDir))
	L.SetGlobal("tar_file", lua.LString(tarFile))

	err := L.DoString(`
		local tar = require("stdlib.tar")
		local err = tar.create(tar_file, {src_dir})
		assert(err == nil, "create should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCreateInvalidPath(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local tar = require("stdlib.tar")
		local err = tar.create("/nonexistent/directory/archive.tar", {"file.txt"})
		assert(err ~= nil, "should error for invalid output path")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// extract tests
// =============================================================================

func TestExtractAll(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	tarFile := filepath.Join(tmpDir, "archive.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	os.WriteFile(srcFile, []byte("Hello, World!"), 0644)

	L.SetGlobal("src_file", lua.LString(srcFile))
	L.SetGlobal("tar_file", lua.LString(tarFile))
	L.SetGlobal("dest_dir", lua.LString(destDir))

	err := L.DoString(`
		local tar = require("stdlib.tar")
		-- First create the tar
		local err = tar.create(tar_file, {src_file})
		assert(err == nil, "create should not error")
		
		-- Then extract it
		err = tar.extract(tar_file, dest_dir)
		assert(err == nil, "extract should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExtractInvalidTar(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	os.WriteFile(invalidFile, []byte("not a tar file"), 0644)

	L.SetGlobal("invalid_file", lua.LString(invalidFile))
	L.SetGlobal("dest_dir", lua.LString(destDir))

	err := L.DoString(`
		local tar = require("stdlib.tar")
		local err = tar.extract(invalid_file, dest_dir)
		assert(err ~= nil, "should error for invalid tar file")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExtractNotFound(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local tar = require("stdlib.tar")
		local err = tar.extract("/nonexistent/archive.tar", "/tmp/dest")
		assert(err ~= nil, "should error for missing file")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list tests
// =============================================================================

func TestListContents(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	tarFile := filepath.Join(tmpDir, "archive.tar")

	os.WriteFile(file1, []byte("Content 1"), 0644)
	os.WriteFile(file2, []byte("Content 2"), 0644)

	L.SetGlobal("file1", lua.LString(file1))
	L.SetGlobal("file2", lua.LString(file2))
	L.SetGlobal("tar_file", lua.LString(tarFile))

	err := L.DoString(`
		local tar = require("stdlib.tar")
		local err = tar.create(tar_file, {file1, file2})
		assert(err == nil, "create should not error")
		
		local files, err = tar.list(tar_file)
		assert(err == nil, "list should not error: " .. tostring(err))
		assert(#files == 2, "should list 2 files")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestListInvalidTar(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local tar = require("stdlib.tar")
		local files, err = tar.list("/nonexistent/archive.tar")
		assert(files == nil, "files should be nil for missing archive")
		assert(err ~= nil, "should error for missing file")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// round trip tests
// =============================================================================

func TestRoundTrip(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "original.txt")
	tarFile := filepath.Join(tmpDir, "archive.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	originalContent := "This is the original content for round trip test."
	os.WriteFile(srcFile, []byte(originalContent), 0644)

	L.SetGlobal("src_file", lua.LString(srcFile))
	L.SetGlobal("tar_file", lua.LString(tarFile))
	L.SetGlobal("dest_dir", lua.LString(destDir))

	err := L.DoString(`
		local tar = require("stdlib.tar")
		
		-- Create tar
		local err = tar.create(tar_file, {src_file})
		assert(err == nil, "create should not error")
		
		-- List contents
		local files, err = tar.list(tar_file)
		assert(err == nil, "list should not error")
		assert(#files == 1, "should have 1 file")
		
		-- Extract
		err = tar.extract(tar_file, dest_dir)
		assert(err == nil, "extract should not error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
