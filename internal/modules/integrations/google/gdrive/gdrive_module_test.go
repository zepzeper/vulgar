package gdrive

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestLoader(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	if err := L.DoString(`local gdrive = require("integrations.gdrive")`); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}
}

func TestExportsExist(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	exports := []string{
		"configure",
		"list_files",
		"get_file",
		"download",
		"upload",
		"delete",
		"create_folder",
		"move",
		"copy",
		"share",
		"rename",
		"search",
	}

	for _, export := range exports {
		if err := L.DoString(`
			local gdrive = require("integrations.gdrive")
			if type(gdrive.` + export + `) ~= "function" then
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
		local gdrive = require("integrations.gdrive")
		local client, err = gdrive.configure({})
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

func TestStringReader(t *testing.T) {
	content := "Hello, World!"
	reader := &stringReader{content: content}

	buf := make([]byte, 5)

	// First read
	n, err := reader.Read(buf)
	if err != nil {
		t.Fatalf("First read failed: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes, got %d", n)
	}
	if string(buf) != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", string(buf))
	}

	// Second read
	n, err = reader.Read(buf)
	if err != nil {
		t.Fatalf("Second read failed: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes, got %d", n)
	}
	if string(buf) != ", Wor" {
		t.Errorf("Expected ', Wor', got '%s'", string(buf))
	}
}
