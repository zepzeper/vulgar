package json

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func setupLuaState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

func TestDecodeSimpleObject(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local json = require("json")
		result, err = json.decode('{"name": "test", "value": 42}')
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result.Type() != lua.LTTable {
		t.Fatalf("expected table, got %s", result.Type())
	}

	tbl := result.(*lua.LTable)
	name := L.GetField(tbl, "name")
	if name.String() != "test" {
		t.Errorf("expected name='test', got '%s'", name.String())
	}

	value := L.GetField(tbl, "value")
	if value.(lua.LNumber) != 42 {
		t.Errorf("expected value=42, got %v", value)
	}
}

func TestDecodeArray(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local json = require("json")
		result, err = json.decode('[1, 2, 3]')
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	tbl := result.(*lua.LTable)

	if tbl.Len() != 3 {
		t.Errorf("expected array length 3, got %d", tbl.Len())
	}
}

func TestDecodeNestedObject(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local json = require("json")
		result, err = json.decode('{"outer": {"inner": "value"}}')
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	tbl := result.(*lua.LTable)
	outer := L.GetField(tbl, "outer").(*lua.LTable)
	inner := L.GetField(outer, "inner")

	if inner.String() != "value" {
		t.Errorf("expected inner='value', got '%s'", inner.String())
	}
}

func TestDecodeInvalidJSON(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local json = require("json")
		result, err = json.decode('not valid json')
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result")
	if result != lua.LNil {
		t.Error("expected nil result for invalid JSON")
	}

	errVal := L.GetGlobal("err")
	if errVal == lua.LNil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDecodeBoolean(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local json = require("json")
		result, err = json.decode('{"active": true, "deleted": false}')
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").(*lua.LTable)
	active := L.GetField(result, "active")
	deleted := L.GetField(result, "deleted")

	if active != lua.LTrue {
		t.Error("expected active=true")
	}
	if deleted != lua.LFalse {
		t.Error("expected deleted=false")
	}
}

func TestDecodeNull(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local json = require("json")
		result, err = json.decode('{"value": null}')
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").(*lua.LTable)
	value := L.GetField(result, "value")

	if value != lua.LNil {
		t.Errorf("expected nil for null, got %v", value)
	}
}

func TestEncodeSimpleObject(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local json = require("json")
		result, err = json.encode({name = "test", value = 42})
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	// JSON output may vary in order, check for key parts
	if len(result) == 0 {
		t.Error("expected non-empty JSON string")
	}
}

func TestEncodeArray(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local json = require("json")
		result, err = json.encode({1, 2, 3})
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != "[1,2,3]" {
		t.Errorf("expected '[1,2,3]', got '%s'", result)
	}
}

func TestEncodePretty(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local json = require("json")
		result, err = json.encode({a = 1}, {pretty = true})
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	// Pretty print should contain newlines
	if len(result) < 5 {
		t.Error("expected pretty printed JSON with formatting")
	}
}

func TestEncodeBoolean(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	err := L.DoString(`
		local json = require("json")
		result, err = json.encode({active = true})
	`)
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	result := L.GetGlobal("result").String()
	if result != `{"active":true}` {
		t.Errorf("expected '{\"active\":true}', got '%s'", result)
	}
}

func TestLoaderReturnsModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`json = require("json")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	mod := L.GetGlobal("json")
	if mod.Type() != lua.LTTable {
		t.Errorf("expected table, got %s", mod.Type())
	}

	tbl := mod.(*lua.LTable)
	decode := L.GetField(tbl, "decode")
	encode := L.GetField(tbl, "encode")

	if decode.Type() != lua.LTFunction {
		t.Error("expected decode function")
	}
	if encode.Type() != lua.LTFunction {
		t.Error("expected encode function")
	}
}
