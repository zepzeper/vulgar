package airtable

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
		local airtable = require("integrations.airtable")
		local client, err = airtable.client({token = "test-token"})
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
		local airtable = require("integrations.airtable")
		local client, err = airtable.client({})
		assert(client == nil or err ~= nil, "should error with missing token")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list_records tests
// =============================================================================

func TestListRecordsNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local airtable = require("integrations.airtable")
		local records, err = airtable.list_records(nil, "base_id", "table_name")
		assert(records == nil, "records should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// get_record tests
// =============================================================================

func TestGetRecordNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local airtable = require("integrations.airtable")
		local record, err = airtable.get_record(nil, "base_id", "table_name", "record_id")
		assert(record == nil, "record should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// create_record tests
// =============================================================================

func TestCreateRecordNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local airtable = require("integrations.airtable")
		local record, err = airtable.create_record(nil, "base_id", "table_name", {
			Name = "Test"
		})
		assert(record == nil, "record should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// update_record tests
// =============================================================================

func TestUpdateRecordNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local airtable = require("integrations.airtable")
		local record, err = airtable.update_record(nil, "base_id", "table_name", "record_id", {
			Name = "Updated"
		})
		assert(record == nil, "record should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// delete_record tests
// =============================================================================

func TestDeleteRecordNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local airtable = require("integrations.airtable")
		local err = airtable.delete_record(nil, "base_id", "table_name", "record_id")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list_bases tests
// =============================================================================

func TestListBasesNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local airtable = require("integrations.airtable")
		local bases, err = airtable.list_bases(nil)
		assert(bases == nil, "bases should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list_tables tests
// =============================================================================

func TestListTablesNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local airtable = require("integrations.airtable")
		local tables, err = airtable.list_tables(nil, "base_id")
		assert(tables == nil, "tables should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
