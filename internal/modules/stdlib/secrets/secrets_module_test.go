package secrets

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
// get tests
// =============================================================================

func TestGetFromEnv(t *testing.T) {
	os.Setenv("TEST_SECRET", "secret_value")
	defer os.Unsetenv("TEST_SECRET")

	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local value, err = secrets.get("TEST_SECRET")
		assert(err == nil, "get should not error: " .. tostring(err))
		assert(value == "secret_value", "should get secret value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestGetMissing(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local value, err = secrets.get("NONEXISTENT_SECRET_XYZ")
		assert(value == nil, "value should be nil for missing secret")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// get_or tests
// =============================================================================

func TestGetOr(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local value = secrets.get_or("NONEXISTENT_XYZ", "default")
		assert(value == "default", "should return default value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestGetOrExists(t *testing.T) {
	os.Setenv("TEST_SECRET", "actual_value")
	defer os.Unsetenv("TEST_SECRET")

	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local value = secrets.get_or("TEST_SECRET", "default")
		assert(value == "actual_value", "should return actual value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// require tests
// =============================================================================

func TestRequireExists(t *testing.T) {
	os.Setenv("REQUIRED_SECRET", "value")
	defer os.Unsetenv("REQUIRED_SECRET")

	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local value, err = secrets.require("REQUIRED_SECRET")
		assert(err == nil, "require should not error: " .. tostring(err))
		assert(value == "value", "should return value")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestRequireMissing(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local value, err = secrets.require("NONEXISTENT_SECRET_XYZ")
		assert(value == nil, "value should be nil")
		assert(err ~= nil, "should error for missing required secret")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// load_env tests
// =============================================================================

func TestLoadEnv(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(envFile, []byte("TEST_VAR=test_value\n"), 0644)

	L := newTestState()
	defer L.Close()
	L.SetGlobal("env_file", lua.LString(envFile))

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local err = secrets.load_env(env_file)
		assert(err == nil, "load_env should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestLoadEnvMissing(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local err = secrets.load_env("/nonexistent/path/.env")
		assert(err ~= nil, "should error for missing file")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// mask tests
// =============================================================================

func TestMask(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local masked = secrets.mask("secret123")
		assert(masked ~= nil, "masked should not be nil")
		assert(masked ~= "secret123", "should mask the value")
		assert(string.find(masked, "*") ~= nil, "should contain asterisks")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestMaskEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local masked = secrets.mask("")
		assert(masked ~= nil, "masked should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// from_file tests
// =============================================================================

func TestFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	secretFile := filepath.Join(tmpDir, "secret")
	os.WriteFile(secretFile, []byte("file_secret"), 0644)

	L := newTestState()
	defer L.Close()
	L.SetGlobal("secret_file", lua.LString(secretFile))

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local value, err = secrets.from_file(secret_file)
		assert(err == nil, "from_file should not error: " .. tostring(err))
		assert(value == "file_secret", "should read secret from file")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestFromFileMissing(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local value, err = secrets.from_file("/nonexistent/path/secret")
		assert(value == nil, "value should be nil")
		assert(err ~= nil, "should error for missing file")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// validate tests
// =============================================================================

func TestValidate(t *testing.T) {
	os.Setenv("VAR1", "value1")
	os.Setenv("VAR2", "value2")
	defer os.Unsetenv("VAR1")
	defer os.Unsetenv("VAR2")

	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local missing, err = secrets.validate({"VAR1", "VAR2"})
		assert(err == nil, "validate should not error: " .. tostring(err))
		assert(#missing == 0, "should have no missing secrets")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestValidateMissing(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local secrets = require("stdlib.secrets")
		local missing, err = secrets.validate({"MISSING_VAR_XYZ"})
		assert(#missing > 0, "should have missing secrets")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
