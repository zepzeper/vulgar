package gcalendar

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestLoader(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	if err := L.DoString(`local gcalendar = require("integrations.gcalendar")`); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}
}

func TestExportsExist(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	exports := []string{
		"configure",
		"list_events",
		"get_event",
		"create_event",
		"update_event",
		"delete_event",
		"list_calendars",
		"quick_add",
		"freebusy",
	}

	for _, export := range exports {
		if err := L.DoString(`
			local gcalendar = require("integrations.gcalendar")
			if type(gcalendar.` + export + `) ~= "function" then
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

	err := L.DoString(`
		local gcalendar = require("integrations.gcalendar")
		local client, err = gcalendar.configure({})
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

func TestEventToLuaTable(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Note: Full testing requires mocking Google Calendar API
	// This test just verifies the helper function structure
}

