package yaml

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
// decode tests
// =============================================================================

func TestDecodeSimpleString(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode("name: John")
		assert(err == nil, "decode should not error: " .. tostring(err))
		assert(data.name == "John", "name should be John, got: " .. tostring(data.name))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeNestedStructure(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode([[
person:
  name: John
  age: 30
  address:
    city: New York
    country: USA
]])
		assert(err == nil, "decode should not error: " .. tostring(err))
		assert(data.person.name == "John", "name should be John")
		assert(data.person.age == 30, "age should be 30")
		assert(data.person.address.city == "New York", "city should be New York")
		assert(data.person.address.country == "USA", "country should be USA")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeArray(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode([[
fruits:
  - apple
  - banana
  - cherry
]])
		assert(err == nil, "decode should not error: " .. tostring(err))
		assert(#data.fruits == 3, "should have 3 fruits")
		assert(data.fruits[1] == "apple", "first fruit should be apple")
		assert(data.fruits[2] == "banana", "second fruit should be banana")
		assert(data.fruits[3] == "cherry", "third fruit should be cherry")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeArrayOfObjects(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode([[
users:
  - name: Alice
    role: admin
  - name: Bob
    role: user
]])
		assert(err == nil, "decode should not error: " .. tostring(err))
		assert(#data.users == 2, "should have 2 users")
		assert(data.users[1].name == "Alice", "first user should be Alice")
		assert(data.users[1].role == "admin", "Alice should be admin")
		assert(data.users[2].name == "Bob", "second user should be Bob")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeBoolean(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode([[
enabled: true
disabled: false
]])
		assert(err == nil, "decode should not error: " .. tostring(err))
		assert(data.enabled == true, "enabled should be true")
		assert(data.disabled == false, "disabled should be false")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeNumbers(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode([[
integer: 42
float: 3.14
negative: -100
]])
		assert(err == nil, "decode should not error: " .. tostring(err))
		assert(data.integer == 42, "integer should be 42")
		assert(data.float == 3.14, "float should be 3.14")
		assert(data.negative == -100, "negative should be -100")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeNull(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode([[
value: null
empty: ~
]])
		assert(err == nil, "decode should not error: " .. tostring(err))
		assert(data.value == nil, "value should be nil")
		assert(data.empty == nil, "empty should be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeMultilineString(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode([[
description: |
  This is a multiline
  string that preserves
  newlines.
]])
		assert(err == nil, "decode should not error: " .. tostring(err))
		assert(data.description ~= nil, "description should not be nil")
		assert(string.find(data.description, "multiline") ~= nil, "should contain multiline")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeInvalidYAML(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode("invalid: [unclosed")
		assert(data == nil, "data should be nil on error")
		assert(err ~= nil, "should return error for invalid yaml")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeEmptyString(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode("")
		assert(err == nil, "decode of empty string should not error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// encode tests
// =============================================================================

func TestEncodeSimpleTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local result, err = yaml.encode({name = "John", age = 30})
		assert(err == nil, "encode should not error: " .. tostring(err))
		assert(result ~= nil, "result should not be nil")
		assert(string.find(result, "name") ~= nil, "should contain 'name'")
		assert(string.find(result, "John") ~= nil, "should contain 'John'")
		assert(string.find(result, "age") ~= nil, "should contain 'age'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEncodeNestedTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local result, err = yaml.encode({
			person = {
				name = "John",
				address = {city = "NYC"}
			}
		})
		assert(err == nil, "encode should not error: " .. tostring(err))
		assert(string.find(result, "person") ~= nil, "should contain 'person'")
		assert(string.find(result, "address") ~= nil, "should contain 'address'")
		assert(string.find(result, "NYC") ~= nil, "should contain 'NYC'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEncodeArray(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local result, err = yaml.encode({items = {"apple", "banana", "cherry"}})
		assert(err == nil, "encode should not error: " .. tostring(err))
		assert(string.find(result, "apple") ~= nil, "should contain 'apple'")
		assert(string.find(result, "banana") ~= nil, "should contain 'banana'")
		assert(string.find(result, "cherry") ~= nil, "should contain 'cherry'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEncodeBoolean(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local result, err = yaml.encode({enabled = true, disabled = false})
		assert(err == nil, "encode should not error: " .. tostring(err))
		assert(string.find(result, "enabled") ~= nil, "should contain 'enabled'")
		assert(string.find(result, "true") ~= nil, "should contain 'true'")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEncodeEmptyTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local result, err = yaml.encode({})
		assert(err == nil, "encode of empty table should not error")
		assert(result ~= nil, "result should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local original = {
			name = "Test",
			count = 42,
			active = true,
			tags = {"a", "b", "c"}
		}
		local encoded, err = yaml.encode(original)
		assert(err == nil, "encode should not error")
		
		local decoded, err = yaml.decode(encoded)
		assert(err == nil, "decode should not error")
		assert(decoded.name == original.name, "name should match")
		assert(decoded.count == original.count, "count should match")
		assert(decoded.active == original.active, "active should match")
		assert(#decoded.tags == #original.tags, "tags length should match")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// decode_file tests
// =============================================================================

func TestDecodeFileSuccess(t *testing.T) {
	L := newTestState()
	defer L.Close()

	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")
	content := []byte("name: Test\nvalue: 123\n")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode_file(test_file)
		assert(err == nil, "decode_file should not error: " .. tostring(err))
		assert(data.name == "Test", "name should be Test")
		assert(data.value == 123, "value should be 123")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeFileNotFound(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode_file("/nonexistent/path/file.yaml")
		assert(data == nil, "data should be nil when file not found")
		assert(err ~= nil, "should return error for missing file")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeFileInvalidYAML(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")
	content := []byte("invalid: [unclosed bracket")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode_file(test_file)
		assert(data == nil, "data should be nil for invalid yaml")
		assert(err ~= nil, "should return error for invalid yaml file")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDecodeFileComplexStructure(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "complex.yaml")
	content := []byte(`
database:
  host: localhost
  port: 5432
  credentials:
    user: admin
    password: secret
servers:
  - name: web1
    port: 8080
  - name: web2
    port: 8081
`)
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode_file(test_file)
		assert(err == nil, "decode_file should not error: " .. tostring(err))
		assert(data.database.host == "localhost", "host should be localhost")
		assert(data.database.port == 5432, "port should be 5432")
		assert(data.database.credentials.user == "admin", "user should be admin")
		assert(#data.servers == 2, "should have 2 servers")
		assert(data.servers[1].name == "web1", "first server should be web1")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// encode_file tests
// =============================================================================

func TestEncodeFileSuccess(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "output.yaml")

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data = {name = "Test", value = 42}
		local err = yaml.encode_file(test_file, data)
		assert(err == nil, "encode_file should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}

	// Verify file contents
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("output file is empty")
	}
}

func TestEncodeFileOverwrite(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "output.yaml")

	// Create initial file
	if err := os.WriteFile(tmpFile, []byte("old: content\n"), 0644); err != nil {
		t.Fatalf("failed to create initial file: %v", err)
	}

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data = {new = "content"}
		local err = yaml.encode_file(test_file, data)
		assert(err == nil, "encode_file should not error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}

	// Verify file was overwritten
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	contentStr := string(content)
	if contentStr == "old: content\n" {
		t.Fatal("file was not overwritten")
	}
}

func TestEncodeFileInvalidPath(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data = {name = "Test"}
		local err = yaml.encode_file("/nonexistent/directory/file.yaml", data)
		assert(err ~= nil, "should return error for invalid path")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestEncodeFileComplexStructure(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "complex.yaml")

	L.SetGlobal("test_file", lua.LString(tmpFile))

	err := L.DoString(`
		local yaml = require("stdlib.yaml")
		local data = {
			database = {
				host = "localhost",
				port = 5432
			},
			servers = {
				{name = "web1", port = 8080},
				{name = "web2", port = 8081}
			}
		}
		local err = yaml.encode_file(test_file, data)
		assert(err == nil, "encode_file should not error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}

	// Read back and verify
	L2 := newTestState()
	defer L2.Close()
	L2.SetGlobal("test_file", lua.LString(tmpFile))

	err = L2.DoString(`
		local yaml = require("stdlib.yaml")
		local data, err = yaml.decode_file(test_file)
		assert(err == nil, "should be able to read back the file")
		assert(data.database.host == "localhost", "host should match")
		assert(#data.servers == 2, "should have 2 servers")
	`)
	if err != nil {
		t.Fatalf("verification failed: %v", err)
	}
}
