package uuid

import (
	"regexp"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func setupLuaState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func TestNewGeneratesValidUUID(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local uuid = require("uuid")
		result = uuid.new()
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if !uuidRegex.MatchString(result) {
		t.Errorf("invalid UUID format: %s", result)
	}
}

func TestNewGeneratesUniqueUUIDs(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local uuid = require("uuid")
		uuid1 = uuid.new()
		uuid2 = uuid.new()
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	uuid1 := L.GetGlobal("uuid1").String()
	uuid2 := L.GetGlobal("uuid2").String()

	if uuid1 == uuid2 {
		t.Error("expected unique UUIDs, got identical values")
	}
}

func TestV4IsAliasForNew(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local uuid = require("uuid")
		result = uuid.v4()
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if !uuidRegex.MatchString(result) {
		t.Errorf("v4() should generate valid UUID: %s", result)
	}
}

func TestParseValidUUID(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local uuid = require("uuid")
		result, err = uuid.parse("550e8400-e29b-41d4-a716-446655440000")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	errVal := L.GetGlobal("err")

	if result != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected parsed UUID, got %s", result)
	}
	if errVal != lua.LNil {
		t.Error("expected no error for valid UUID")
	}
}

func TestParseInvalidUUID(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local uuid = require("uuid")
		result, err = uuid.parse("not-a-uuid")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	errVal := L.GetGlobal("err")

	if result != lua.LNil {
		t.Error("expected nil result for invalid UUID")
	}
	if errVal == lua.LNil {
		t.Error("expected error for invalid UUID")
	}
}

func TestParseEmptyString(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local uuid = require("uuid")
		result, err = uuid.parse("")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result != lua.LNil {
		t.Error("expected nil result for empty string")
	}
}

func TestIsValidWithValidUUID(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local uuid = require("uuid")
		result = uuid.is_valid("550e8400-e29b-41d4-a716-446655440000")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result != lua.LTrue {
		t.Error("expected true for valid UUID")
	}
}

func TestIsValidWithInvalidUUID(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local uuid = require("uuid")
		result = uuid.is_valid("invalid")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result != lua.LFalse {
		t.Error("expected false for invalid UUID")
	}
}

func TestIsValidWithEmptyString(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local uuid = require("uuid")
		result = uuid.is_valid("")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result != lua.LFalse {
		t.Error("expected false for empty string")
	}
}

func TestVersionReturnsCorrectVersion(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	// UUID v4 should return version 4
	err := L.DoString(`
		local uuid = require("uuid")
		local id = uuid.new()
		result, err = uuid.version(id)
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").(lua.LNumber)
	if result != 4 {
		t.Errorf("expected version 4, got %v", result)
	}
}

func TestVersionWithInvalidUUID(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local uuid = require("uuid")
		result, err = uuid.version("invalid")
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	errVal := L.GetGlobal("err")

	if result != lua.LNil {
		t.Error("expected nil result for invalid UUID")
	}
	if errVal == lua.LNil {
		t.Error("expected error for invalid UUID")
	}
}

func TestLoaderReturnsModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`uuid = require("uuid")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	mod := L.GetGlobal("uuid")
	if mod.Type() != lua.LTTable {
		t.Errorf("expected table, got %s", mod.Type())
	}
}
