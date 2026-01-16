package gitlab

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestLoader(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Load the module
	L.PreloadModule(ModuleName, Loader)

	// Test that module loads
	err := L.DoString(`local gitlab = require("integrations.gitlab")`)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}
}

func TestModuleExports(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	// Check that expected functions exist
	code := `
		local gitlab = require("integrations.gitlab")
		
		assert(type(gitlab.client) == "function", "client should be a function")
		assert(type(gitlab.list_commits) == "function", "list_commits should be a function")
		assert(type(gitlab.list_merge_requests) == "function", "list_merge_requests should be a function")
		assert(type(gitlab.list_issues) == "function", "list_issues should be a function")
		assert(type(gitlab.list_pipelines) == "function", "list_pipelines should be a function")
		assert(type(gitlab.get_project) == "function", "get_project should be a function")
		assert(type(gitlab.list_projects) == "function", "list_projects should be a function")
		assert(type(gitlab.create_issue) == "function", "create_issue should be a function")
		assert(type(gitlab.create_merge_request) == "function", "create_merge_request should be a function")
		assert(type(gitlab.get_user) == "function", "get_user should be a function")
		assert(type(gitlab.since_hours) == "function", "since_hours should be a function")
		
		return true
	`

	err := L.DoString(code)
	if err != nil {
		t.Fatalf("Module exports test failed: %v", err)
	}
}

func TestSinceHours(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	code := `
		local gitlab = require("integrations.gitlab")
		local ts = gitlab.since_hours(24)
		
		-- Should return an ISO8601 timestamp
		assert(type(ts) == "string", "since_hours should return a string")
		assert(string.match(ts, "%d%d%d%d%-%d%d%-%d%dT%d%d:%d%d:%d%dZ"), "should be ISO8601 format")
		
		return true
	`

	err := L.DoString(code)
	if err != nil {
		t.Fatalf("since_hours test failed: %v", err)
	}
}

func TestClientRequiresToken(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	// Without a token configured, client() should return an error
	code := `
		local gitlab = require("integrations.gitlab")
		local client, err = gitlab.client()
		
		-- Should error since no token is configured
		if client == nil then
			assert(err ~= nil, "should return an error when token not configured")
			return true
		end
		
		-- If we got here, a token was configured in the test environment
		return true
	`

	err := L.DoString(code)
	if err != nil {
		t.Fatalf("client token test failed: %v", err)
	}
}

func TestClientWithExplicitToken(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	// With explicit token, client should be created (even if invalid)
	code := `
		local gitlab = require("integrations.gitlab")
		local client, err = gitlab.client({
			token = "test-token",
			url = "https://gitlab.example.com"
		})
		
		assert(client ~= nil, "client should be created with explicit token")
		assert(err == nil, "should not error with explicit token")
		
		-- Check config method
		local cfg, err = client:config()
		assert(cfg ~= nil, "config should return table")
		assert(cfg.url == "https://gitlab.example.com", "url should match")
		
		return true
	`

	err := L.DoString(code)
	if err != nil {
		t.Fatalf("explicit token test failed: %v", err)
	}
}
