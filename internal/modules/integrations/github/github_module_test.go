package github

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestLoader(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	err := L.DoString(`local github = require("integrations.github")`)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}
}

func TestModuleExports(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	code := `
		local github = require("integrations.github")
		
		assert(type(github.client) == "function", "client should be a function")
		assert(type(github.get_repo) == "function", "get_repo should be a function")
		assert(type(github.list_repos) == "function", "list_repos should be a function")
		assert(type(github.list_issues) == "function", "list_issues should be a function")
		assert(type(github.list_prs) == "function", "list_prs should be a function")
		assert(type(github.list_commits) == "function", "list_commits should be a function")
		assert(type(github.create_issue) == "function", "create_issue should be a function")
		assert(type(github.create_pr) == "function", "create_pr should be a function")
		assert(type(github.get_user) == "function", "get_user should be a function")
		
		return true
	`

	err := L.DoString(code)
	if err != nil {
		t.Fatalf("Module exports test failed: %v", err)
	}
}

func TestClientRequiresToken(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)

	code := `
		local github = require("integrations.github")
		local client, err = github.client()
		
		if client == nil then
			assert(err ~= nil, "should return an error when token not configured")
			return true
		end
		
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

	code := `
		local github = require("integrations.github")
		local client, err = github.client({
			token = "test-token"
		})
		
		assert(client ~= nil, "client should be created with explicit token")
		assert(err == nil, "should not error with explicit token")
		
		return true
	`

	err := L.DoString(code)
	if err != nil {
		t.Fatalf("explicit token test failed: %v", err)
	}
}
