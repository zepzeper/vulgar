package graphql

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

// =============================================================================
// query tests
// =============================================================================

func TestQueryBasic(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local graphql = require("integrations.graphql")
		
		-- Using public Countries GraphQL API
		local result, err = graphql.query("https://countries.trevorblades.com/", [[
			query {
				countries {
					code
					name
				}
			}
		]])
		
		assert(err == nil, "query should not error: " .. tostring(err))
		assert(result ~= nil, "result should not be nil")
		assert(result.data ~= nil, "should have data")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQueryWithVariables(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local graphql = require("integrations.graphql")
		
		local result, err = graphql.query("https://countries.trevorblades.com/", [[
			query($code: ID!) {
				country(code: $code) {
					name
					capital
				}
			}
		]], {variables = {code = "US"}})
		
		assert(err == nil, "query with vars should not error: " .. tostring(err))
		assert(result ~= nil, "result should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQueryInvalidURL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local graphql = require("integrations.graphql")
		
		local result, err = graphql.query("not-a-url", "query { test }")
		assert(result == nil, "result should be nil for invalid URL")
		assert(err ~= nil, "should error for invalid URL")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQueryWithHeaders(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local graphql = require("integrations.graphql")
		
		local result, err = graphql.query("https://countries.trevorblades.com/", [[
			query { countries { code } }
		]], {
			headers = {
				["X-Custom"] = "value"
			}
		})
		
		assert(err == nil, "query with headers should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// mutate tests
// =============================================================================

func TestMutateNoServer(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local graphql = require("integrations.graphql")
		
		local result, err = graphql.mutate("https://invalid-graphql-server.test", [[
			mutation { test }
		]])
		-- Should fail to connect
		assert(result == nil or err ~= nil, "should error for invalid server")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// subscribe tests
// =============================================================================

func TestSubscribeNoServer(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local graphql = require("integrations.graphql")
		
		local sub, err = graphql.subscribe("wss://invalid-graphql-server.test", [[
			subscription { test }
		]], function() end)
		-- Should fail
		assert(sub == nil or err ~= nil, "should error for invalid server")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// client tests
// =============================================================================

func TestClientCreate(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local graphql = require("integrations.graphql")
		
		local client, err = graphql.client("https://countries.trevorblades.com/", {
			headers = {["Authorization"] = "Bearer token"}
		})
		assert(err == nil, "client should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestClientQuery(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local graphql = require("integrations.graphql")
		
		local client, _ = graphql.client("https://countries.trevorblades.com/")
		local result, err = client:query([[
			query { countries { code } }
		]])
		
		assert(err == nil, "client query should not error: " .. tostring(err))
		assert(result ~= nil, "result should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// introspect tests
// =============================================================================

func TestIntrospect(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local graphql = require("integrations.graphql")
		
		local schema, err = graphql.introspect("https://countries.trevorblades.com/")
		assert(err == nil, "introspect should not error: " .. tostring(err))
		assert(schema ~= nil, "schema should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestIntrospectInvalidServer(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local graphql = require("integrations.graphql")
		
		local schema, err = graphql.introspect("https://invalid-server.test")
		assert(schema == nil or err ~= nil, "should error for invalid server")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
