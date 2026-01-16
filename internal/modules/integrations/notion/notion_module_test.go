package notion

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
// client tests
// =============================================================================

func TestClientCreate(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local notion = require("integrations.notion")
		local client, err = notion.client({token = "test-token"})
		assert(err == nil, "client should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestClientMissingToken(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local notion = require("integrations.notion")
		local client, err = notion.client({})
		assert(client == nil or err ~= nil, "should error with missing token")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// get_page tests
// =============================================================================

func TestGetPageNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local notion = require("integrations.notion")
		local page, err = notion.get_page(nil, "page_id")
		assert(page == nil, "page should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// create_page tests
// =============================================================================

func TestCreatePageNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local notion = require("integrations.notion")
		local page, err = notion.create_page(nil, {
			parent = {database_id = "db_id"},
			properties = {}
		})
		assert(page == nil, "page should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// update_page tests
// =============================================================================

func TestUpdatePageNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local notion = require("integrations.notion")
		local page, err = notion.update_page(nil, "page_id", {properties = {}})
		assert(page == nil, "page should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// query_database tests
// =============================================================================

func TestQueryDatabaseNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local notion = require("integrations.notion")
		local results, err = notion.query_database(nil, "database_id", {})
		assert(results == nil, "results should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// get_database tests
// =============================================================================

func TestGetDatabaseNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local notion = require("integrations.notion")
		local db, err = notion.get_database(nil, "database_id")
		assert(db == nil, "db should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// append_blocks tests
// =============================================================================

func TestAppendBlocksNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local notion = require("integrations.notion")
		local err = notion.append_blocks(nil, "page_id", {})
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// get_blocks tests
// =============================================================================

func TestGetBlocksNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local notion = require("integrations.notion")
		local blocks, err = notion.get_blocks(nil, "page_id")
		assert(blocks == nil, "blocks should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// search tests
// =============================================================================

func TestSearchNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local notion = require("integrations.notion")
		local results, err = notion.search(nil, "query")
		assert(results == nil, "results should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
