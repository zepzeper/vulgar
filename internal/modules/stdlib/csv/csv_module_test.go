package csv

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
// parse tests
// =============================================================================

func TestParseSimpleCSV(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local rows, err = csv.parse("a,b,c\n1,2,3")
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(#rows == 2, "should have 2 rows")
		assert(rows[1][1] == "a", "first cell should be 'a'")
		assert(rows[2][1] == "1", "second row first cell should be '1'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseWithHeader(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local rows, err = csv.parse("name,age\nJohn,30\nJane,25", {header = true})
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(#rows == 2, "should have 2 data rows")
		assert(rows[1].name == "John", "first row name should be John")
		assert(rows[1].age == "30", "first row age should be 30")
		assert(rows[2].name == "Jane", "second row name should be Jane")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseCustomDelimiter(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local rows, err = csv.parse("a;b;c\n1;2;3", {delimiter = ";"})
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(#rows == 2, "should have 2 rows")
		assert(rows[1][2] == "b", "second cell should be 'b'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseQuotedFields(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local rows, err = csv.parse('"hello, world",test\nfoo,bar')
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(rows[1][1] == "hello, world", "quoted field should preserve comma")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseEmptyCSV(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local rows, err = csv.parse("")
		assert(err == nil, "parse empty should not error")
		assert(#rows == 0, "should have 0 rows")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestParseEmptyFields(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local rows, err = csv.parse("a,,c\n1,2,")
		assert(err == nil, "parse should not error: " .. tostring(err))
		assert(rows[1][2] == "", "empty field should be empty string")
		assert(rows[2][3] == "", "trailing empty field should be empty string")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// encode tests
// =============================================================================

func TestEncodeSimpleRows(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local result, err = csv.encode({{"a", "b", "c"}, {"1", "2", "3"}})
		assert(err == nil, "encode should not error: " .. tostring(err))
		assert(result ~= nil, "result should not be nil")
		assert(string.find(result, "a,b,c") ~= nil, "should contain first row")
		assert(string.find(result, "1,2,3") ~= nil, "should contain second row")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEncodeWithHeader(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local rows = {
			{name = "John", age = "30"},
			{name = "Jane", age = "25"}
		}
		local result, err = csv.encode(rows, {header = true})
		assert(err == nil, "encode should not error: " .. tostring(err))
		assert(string.find(result, "name") ~= nil, "should contain header 'name'")
		assert(string.find(result, "age") ~= nil, "should contain header 'age'")
		assert(string.find(result, "John") ~= nil, "should contain 'John'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEncodeCustomDelimiter(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local result, err = csv.encode({{"a", "b", "c"}}, {delimiter = ";"})
		assert(err == nil, "encode should not error: " .. tostring(err))
		assert(string.find(result, "a;b;c") ~= nil, "should use semicolon delimiter")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEncodeQuotesFieldsWithComma(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local result, err = csv.encode({{"hello, world", "test"}})
		assert(err == nil, "encode should not error: " .. tostring(err))
		assert(string.find(result, '"hello, world"') ~= nil, "should quote field with comma")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEncodeEmptyTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local result, err = csv.encode({})
		assert(err == nil, "encode empty should not error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// read_file tests
// =============================================================================

func TestReadFileSuccess(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.csv")
	content := []byte("name,age\nJohn,30\nJane,25")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local rows, err = csv.read_file(test_file, {header = true})
		assert(err == nil, "read_file should not error: " .. tostring(err))
		assert(#rows == 2, "should have 2 data rows")
		assert(rows[1].name == "John", "first row name should be John")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestReadFileNotFound(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local rows, err = csv.read_file("/nonexistent/file.csv")
		assert(rows == nil, "rows should be nil when file not found")
		assert(err ~= nil, "should return error for missing file")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestReadFileWithOptions(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.csv")
	content := []byte("a;b;c\n1;2;3")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local rows, err = csv.read_file(test_file, {delimiter = ";"})
		assert(err == nil, "read_file should not error: " .. tostring(err))
		assert(rows[1][1] == "a", "should parse with custom delimiter")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// write_file tests
// =============================================================================

func TestWriteFileSuccess(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "output.csv")

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local rows = {{"a", "b", "c"}, {"1", "2", "3"}}
		local err = csv.write_file(test_file, rows)
		assert(err == nil, "write_file should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("output file is empty")
	}
}

func TestWriteFileWithHeader(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "output.csv")

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local rows = {
			{name = "John", age = "30"},
			{name = "Jane", age = "25"}
		}
		local err = csv.write_file(test_file, rows, {header = true})
		assert(err == nil, "write_file should not error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	contentStr := string(content)
	if contentStr == "" {
		t.Fatal("output file is empty")
	}
}

func TestWriteFileInvalidPath(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local rows = {{"a", "b"}}
		local err = csv.write_file("/nonexistent/directory/file.csv", rows)
		assert(err ~= nil, "should return error for invalid path")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestWriteFileRoundTrip(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "roundtrip.csv")

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local csv = require("stdlib.csv")
		local original = {{"name", "age"}, {"John", "30"}, {"Jane", "25"}}
		
		local err = csv.write_file(test_file, original)
		assert(err == nil, "write should not error")
		
		local rows, err = csv.read_file(test_file)
		assert(err == nil, "read should not error")
		assert(#rows == 3, "should have 3 rows")
		assert(rows[1][1] == "name", "first cell should match")
		assert(rows[2][1] == "John", "data should match")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
