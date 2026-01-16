package env

import (
	"os"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func setupLuaState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

func TestGetExistingVar(t *testing.T) {
	_ = os.Setenv("TEST_VAR", "test_value")
	defer func() { _ = os.Unsetenv("TEST_VAR") }()

	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		result = env.get("TEST_VAR")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != "test_value" {
		t.Errorf("expected 'test_value', got '%s'", result)
	}
}

func TestGetMissingVarReturnsNil(t *testing.T) {
	_ = os.Unsetenv("MISSING_VAR_FOR_TEST")

	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		result = env.get("MISSING_VAR_FOR_TEST")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result != lua.LNil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestGetMissingVarWithDefault(t *testing.T) {
	_ = os.Unsetenv("MISSING_VAR_FOR_TEST")

	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		result = env.get("MISSING_VAR_FOR_TEST", "default_value")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != "default_value" {
		t.Errorf("expected 'default_value', got '%s'", result)
	}
}

func TestGetMissingVarWithEmptyDefault(t *testing.T) {
	_ = os.Unsetenv("MISSING_VAR_FOR_TEST")

	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		result = env.get("MISSING_VAR_FOR_TEST", "")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestGetExistingVarIgnoresDefault(t *testing.T) {
	_ = os.Setenv("TEST_VAR", "actual_value")
	defer func() { _ = os.Unsetenv("TEST_VAR") }()

	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		result = env.get("TEST_VAR", "default_value")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != "actual_value" {
		t.Errorf("expected 'actual_value', got '%s'", result)
	}
}

func TestSet(t *testing.T) {
	defer func() { _ = os.Unsetenv("NEW_TEST_VAR") }()

	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		err = env.set("NEW_TEST_VAR", "new_value")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	errVal := L.GetGlobal("err")
	if errVal != lua.LNil {
		t.Errorf("expected no error, got %v", errVal)
	}

	value := os.Getenv("NEW_TEST_VAR")
	if value != "new_value" {
		t.Errorf("expected 'new_value', got '%s'", value)
	}
}

func TestSetOverwrite(t *testing.T) {
	_ = os.Setenv("OVERWRITE_VAR", "old_value")
	defer func() { _ = os.Unsetenv("OVERWRITE_VAR") }()

	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		env.set("OVERWRITE_VAR", "new_value")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	value := os.Getenv("OVERWRITE_VAR")
	if value != "new_value" {
		t.Errorf("expected 'new_value', got '%s'", value)
	}
}

func TestExistsTrue(t *testing.T) {
	_ = os.Setenv("EXISTS_VAR", "value")
	defer func() { _ = os.Unsetenv("EXISTS_VAR") }()

	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		result = env.exists("EXISTS_VAR")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result != lua.LTrue {
		t.Error("expected true")
	}
}

func TestExistsFalse(t *testing.T) {
	_ = os.Unsetenv("MISSING_EXISTS_VAR")

	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		result = env.exists("MISSING_EXISTS_VAR")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result != lua.LFalse {
		t.Error("expected false")
	}
}

func TestExistsEmptyValue(t *testing.T) {
	_ = os.Setenv("EMPTY_VAR", "")
	defer func() { _ = os.Unsetenv("EMPTY_VAR") }()

	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		result = env.exists("EMPTY_VAR")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result != lua.LTrue {
		t.Error("expected true for empty but existing var")
	}
}

func TestUnset(t *testing.T) {
	_ = os.Setenv("UNSET_VAR", "value")

	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		err = env.unset("UNSET_VAR")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	_, exists := os.LookupEnv("UNSET_VAR")
	if exists {
		t.Error("expected var to be unset")
	}
}

func TestUnsetNonexistent(t *testing.T) {
	_ = os.Unsetenv("NONEXISTENT_UNSET_VAR")

	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		err = env.unset("NONEXISTENT_UNSET_VAR")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	errVal := L.GetGlobal("err")
	if errVal != lua.LNil {
		t.Errorf("expected no error for unsetting nonexistent var, got %v", errVal)
	}
}

func TestAll(t *testing.T) {
	_ = os.Setenv("ALL_TEST_VAR", "all_test_value")
	defer func() { _ = os.Unsetenv("ALL_TEST_VAR") }()

	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		result = env.all()
		test_value = result["ALL_TEST_VAR"]
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result.Type() != lua.LTTable {
		t.Fatalf("expected table, got %s", result.Type())
	}

	testValue := L.GetGlobal("test_value").String()
	if testValue != "all_test_value" {
		t.Errorf("expected 'all_test_value', got '%s'", testValue)
	}
}

func TestAllReturnsTable(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local env = require("env")
		result = env.all()
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result.Type() != lua.LTTable {
		t.Errorf("expected table, got %s", result.Type())
	}
}

func TestLoaderReturnsModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`env = require("env")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	mod := L.GetGlobal("env")
	if mod.Type() != lua.LTTable {
		t.Errorf("expected table, got %s", mod.Type())
	}
}

func TestLoaderExportsAllFunctions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`env = require("env")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	tbl := L.GetGlobal("env").(*lua.LTable)

	funcs := []string{"get", "set", "exists", "unset", "all"}
	for _, name := range funcs {
		fn := L.GetField(tbl, name)
		if fn.Type() != lua.LTFunction {
			t.Errorf("expected %s to be a function", name)
		}
	}
}
