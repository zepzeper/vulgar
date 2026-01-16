package zip

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
	zipFile := filepath.Join(tmpDir, "archive.zip")

	if err := os.WriteFile(srcFile, []byte("Hello, World!"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	L.SetGlobal("src_file", lua.LString(srcFile))
	L.SetGlobal("zip_file", lua.LString(zipFile))

	err := L.DoString(`
		local zip = require("stdlib.zip")
		local err = zip.create(zip_file, {src_file})
		assert(err == nil, "create should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}

	// Verify zip was created
	if _, err := os.Stat(zipFile); os.IsNotExist(err) {
		t.Fatal("zip file was not created")
	}
}

func TestCreateFromMultipleFiles(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	zipFile := filepath.Join(tmpDir, "archive.zip")

	os.WriteFile(file1, []byte("Content 1"), 0644)
	os.WriteFile(file2, []byte("Content 2"), 0644)

	L.SetGlobal("file1", lua.LString(file1))
	L.SetGlobal("file2", lua.LString(file2))
	L.SetGlobal("zip_file", lua.LString(zipFile))

	err := L.DoString(`
		local zip = require("stdlib.zip")
		local err = zip.create(zip_file, {file1, file2})
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

	zipFile := filepath.Join(tmpDir, "archive.zip")

	L.SetGlobal("src_dir", lua.LString(srcDir))
	L.SetGlobal("zip_file", lua.LString(zipFile))

	err := L.DoString(`
		local zip = require("stdlib.zip")
		local err = zip.create(zip_file, {src_dir})
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
		local zip = require("stdlib.zip")
		local err = zip.create("/nonexistent/directory/archive.zip", {"file.txt"})
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
	zipFile := filepath.Join(tmpDir, "archive.zip")
	destDir := filepath.Join(tmpDir, "extracted")

	os.WriteFile(srcFile, []byte("Hello, World!"), 0644)

	L.SetGlobal("src_file", lua.LString(srcFile))
	L.SetGlobal("zip_file", lua.LString(zipFile))
	L.SetGlobal("dest_dir", lua.LString(destDir))

	err := L.DoString(`
		local zip = require("stdlib.zip")
		-- First create the zip
		local err = zip.create(zip_file, {src_file})
		assert(err == nil, "create should not error")
		
		-- Then extract it
		err = zip.extract(zip_file, dest_dir)
		assert(err == nil, "extract should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExtractInvalidZip(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.zip")
	destDir := filepath.Join(tmpDir, "extracted")

	os.WriteFile(invalidFile, []byte("not a zip file"), 0644)

	L.SetGlobal("invalid_file", lua.LString(invalidFile))
	L.SetGlobal("dest_dir", lua.LString(destDir))

	err := L.DoString(`
		local zip = require("stdlib.zip")
		local err = zip.extract(invalid_file, dest_dir)
		assert(err ~= nil, "should error for invalid zip file")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExtractNotFound(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local zip = require("stdlib.zip")
		local err = zip.extract("/nonexistent/archive.zip", "/tmp/dest")
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
	zipFile := filepath.Join(tmpDir, "archive.zip")

	os.WriteFile(file1, []byte("Content 1"), 0644)
	os.WriteFile(file2, []byte("Content 2"), 0644)

	L.SetGlobal("file1", lua.LString(file1))
	L.SetGlobal("file2", lua.LString(file2))
	L.SetGlobal("zip_file", lua.LString(zipFile))

	err := L.DoString(`
		local zip = require("stdlib.zip")
		local err = zip.create(zip_file, {file1, file2})
		assert(err == nil, "create should not error")
		
		local files, err = zip.list(zip_file)
		assert(err == nil, "list should not error: " .. tostring(err))
		assert(#files == 2, "should list 2 files")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestListInvalidZip(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local zip = require("stdlib.zip")
		local files, err = zip.list("/nonexistent/archive.zip")
		assert(files == nil, "files should be nil for missing archive")
		assert(err ~= nil, "should error for missing file")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// add_file tests
// =============================================================================

func TestAddFileToExisting(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	zipFile := filepath.Join(tmpDir, "archive.zip")

	os.WriteFile(file1, []byte("Content 1"), 0644)
	os.WriteFile(file2, []byte("Content 2"), 0644)

	L.SetGlobal("file1", lua.LString(file1))
	L.SetGlobal("file2", lua.LString(file2))
	L.SetGlobal("zip_file", lua.LString(zipFile))

	err := L.DoString(`
		local zip = require("stdlib.zip")
		-- Create initial zip with one file
		local err = zip.create(zip_file, {file1})
		assert(err == nil, "create should not error")
		
		-- Add another file
		err = zip.add_file(zip_file, file2, "added_file.txt")
		assert(err == nil, "add_file should not error: " .. tostring(err))
		
		-- Verify both files are in the archive
		local files, _ = zip.list(zip_file)
		assert(#files == 2, "should have 2 files after adding")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestAddFileInvalidArchive(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(srcFile, []byte("Content"), 0644)

	L.SetGlobal("src_file", lua.LString(srcFile))

	err := L.DoString(`
		local zip = require("stdlib.zip")
		local err = zip.add_file("/nonexistent/archive.zip", src_file, "file.txt")
		assert(err ~= nil, "should error for missing archive")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
