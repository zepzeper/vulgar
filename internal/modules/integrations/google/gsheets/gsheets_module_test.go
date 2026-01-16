package gsheets

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

func TestLoader(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	if err := L.DoString(`local gsheets = require("integrations.gsheets")`); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}
}

func TestExportsExist(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	exports := []string{
		"configure",
		"get_values",
		"set_values",
		"append_values",
		"clear_values",
		"get_spreadsheet",
		"create_spreadsheet",
		"add_sheet",
		"delete_sheet",
		"batch_update",
		"batch_get_values",
		"find_row",
	}

	for _, export := range exports {
		if err := L.DoString(`
			local gsheets = require("integrations.gsheets")
			if type(gsheets.` + export + `) ~= "function" then
				error("` + export + ` is not a function")
			end
		`); err != nil {
			t.Errorf("Export %s check failed: %v", export, err)
		}
	}
}

func TestConfigureRequiresCredentials(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	// Should fail without credentials
	err := L.DoString(`
		local gsheets = require("integrations.gsheets")
		local client, err = gsheets.configure({})
		if err == nil then
			error("expected error for missing credentials")
		end
		if not string.find(err, "credentials") then
			error("error should mention credentials: " .. err)
		end
	`)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}
}

func TestLuaTableTo2DSlice(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a Lua table
	table := L.NewTable()

	row1 := L.NewTable()
	row1.RawSetInt(1, lua.LString("Name"))
	row1.RawSetInt(2, lua.LString("Age"))

	row2 := L.NewTable()
	row2.RawSetInt(1, lua.LString("John"))
	row2.RawSetInt(2, lua.LNumber(30))

	table.RawSetInt(1, row1)
	table.RawSetInt(2, row2)

	// Convert to values using util function
	values := util.LuaTableTo2DSlice(table)

	if len(values) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(values))
	}

	if len(values[0]) != 2 {
		t.Fatalf("Expected 2 columns in row 1, got %d", len(values[0]))
	}

	if values[0][0] != "Name" {
		t.Errorf("Expected 'Name', got %v", values[0][0])
	}

	if values[1][1] != float64(30) {
		t.Errorf("Expected 30, got %v", values[1][1])
	}
}

func TestGoToLuaIntegration(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test that util.GoToLua works correctly for Sheets data types
	tests := []struct {
		input    interface{}
		expected lua.LValue
	}{
		{"hello", lua.LString("hello")},
		{float64(42), lua.LNumber(42)},
		{true, lua.LBool(true)},
		{nil, lua.LNil},
	}

	for _, tt := range tests {
		result := util.GoToLua(L, tt.input)
		if result.Type() != tt.expected.Type() {
			t.Errorf("GoToLua(%v): expected type %v, got %v", tt.input, tt.expected.Type(), result.Type())
		}
	}
}

func TestLuaToGoIntegration(t *testing.T) {
	// Test that util.LuaToGo works correctly for Sheets data types
	tests := []struct {
		input    lua.LValue
		expected interface{}
	}{
		{lua.LString("hello"), "hello"},
		{lua.LNumber(42), float64(42)},
		{lua.LBool(true), true},
		{lua.LNil, nil},
	}

	for _, tt := range tests {
		result := util.LuaToGo(tt.input)
		if result != tt.expected {
			t.Errorf("LuaToGo(%v): expected %v, got %v", tt.input, tt.expected, result)
		}
	}
}

// Integration tests - require actual Google credentials
// These are skipped by default and can be run with: go test -tags=integration

func TestIntegration(t *testing.T) {
	t.Skip("Integration tests require Google credentials")

	// Example integration test structure:
	// L := lua.NewState()
	// defer L.Close()
	//
	// L.PreloadModule(ModuleName, Loader)
	//
	// err := L.DoString(`
	//     local gsheets = require("integrations.gsheets")
	//     local client, err = gsheets.configure({
	//         credentials_file = os.getenv("GOOGLE_CREDENTIALS_FILE")
	//     })
	//     if err then error(err) end
	//
	//     local spreadsheet, err = gsheets.create_spreadsheet(client, {
	//         title = "Test Spreadsheet"
	//     })
	//     if err then error(err) end
	//
	//     print("Created spreadsheet: " .. spreadsheet.spreadsheet_id)
	// `)
	// if err != nil {
	//     t.Fatalf("Integration test failed: %v", err)
	// }
}
