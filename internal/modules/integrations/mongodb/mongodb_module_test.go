package mongodb

import (
	"os"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

func skipIfNoMongoDB(t *testing.T) string {
	uri := os.Getenv("MONGODB_TEST_URI")
	if uri == "" {
		t.Skip("MONGODB_TEST_URI not set, skipping integration test")
	}
	return uri
}

// =============================================================================
// connect tests
// =============================================================================

func TestConnectMissingConfig(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local client, err = mongodb.connect({})
		assert(client == nil or err ~= nil, "should error with empty config")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectInvalidURI(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local client, err = mongodb.connect({uri = "invalid-uri"})
		assert(client == nil, "client should be nil for invalid URI")
		assert(err ~= nil, "should error for invalid URI")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectValid(t *testing.T) {
	uri := skipIfNoMongoDB(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local client, err = mongodb.connect({uri = test_uri, database = "test"})
		assert(err == nil, "connect should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
		mongodb.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// find_one tests
// =============================================================================

func TestFindOneNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local doc, err = mongodb.find_one(nil, "collection", {})
		assert(doc == nil, "doc should be nil without connection")
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestFindOne(t *testing.T) {
	uri := skipIfNoMongoDB(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local client, _ = mongodb.connect({uri = test_uri, database = "test"})
		
		-- Insert a test document first
		local result, _ = mongodb.insert_one(client, "test_collection", {name = "test", value = 123})
		
		-- Find it
		local doc, err = mongodb.find_one(client, "test_collection", {name = "test"})
		assert(err == nil, "find_one should not error: " .. tostring(err))
		assert(doc ~= nil, "doc should not be nil")
		assert(doc.name == "test", "name should match")
		
		-- Clean up
		mongodb.delete_one(client, "test_collection", {name = "test"})
		mongodb.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestFindOneNoMatch(t *testing.T) {
	uri := skipIfNoMongoDB(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local client, _ = mongodb.connect({uri = test_uri, database = "test"})
		local doc, err = mongodb.find_one(client, "test_collection", {nonexistent_field = "xyz123"})
		-- Should return nil when no match, not error
		assert(doc == nil, "doc should be nil when no match")
		mongodb.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// find tests
// =============================================================================

func TestFind(t *testing.T) {
	uri := skipIfNoMongoDB(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local client, _ = mongodb.connect({uri = test_uri, database = "test"})
		
		-- Insert test documents
		mongodb.insert_one(client, "find_test", {category = "A", value = 1})
		mongodb.insert_one(client, "find_test", {category = "A", value = 2})
		mongodb.insert_one(client, "find_test", {category = "B", value = 3})
		
		-- Find documents
		local docs, err = mongodb.find(client, "find_test", {category = "A"})
		assert(err == nil, "find should not error: " .. tostring(err))
		assert(#docs == 2, "should find 2 documents")
		
		-- Clean up
		mongodb.delete_many(client, "find_test", {})
		mongodb.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestFindWithOptions(t *testing.T) {
	uri := skipIfNoMongoDB(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local client, _ = mongodb.connect({uri = test_uri, database = "test"})
		
		-- Insert test documents
		for i = 1, 10 do
			mongodb.insert_one(client, "find_options_test", {index = i})
		end
		
		-- Find with limit
		local docs, err = mongodb.find(client, "find_options_test", {}, {limit = 5})
		assert(err == nil, "find should not error")
		assert(#docs == 5, "should return limited results")
		
		-- Clean up
		mongodb.delete_many(client, "find_options_test", {})
		mongodb.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// insert tests
// =============================================================================

func TestInsertOne(t *testing.T) {
	uri := skipIfNoMongoDB(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local client, _ = mongodb.connect({uri = test_uri, database = "test"})
		
		local result, err = mongodb.insert_one(client, "insert_test", {
			name = "John",
			age = 30,
			tags = {"developer", "golang"}
		})
		assert(err == nil, "insert_one should not error: " .. tostring(err))
		assert(result ~= nil, "result should not be nil")
		
		-- Clean up
		mongodb.delete_many(client, "insert_test", {})
		mongodb.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestInsertMany(t *testing.T) {
	uri := skipIfNoMongoDB(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local client, _ = mongodb.connect({uri = test_uri, database = "test"})
		
		local docs = {
			{name = "Alice", age = 25},
			{name = "Bob", age = 30},
			{name = "Charlie", age = 35}
		}
		
		local result, err = mongodb.insert_many(client, "insert_many_test", docs)
		assert(err == nil, "insert_many should not error: " .. tostring(err))
		
		-- Clean up
		mongodb.delete_many(client, "insert_many_test", {})
		mongodb.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// update tests
// =============================================================================

func TestUpdateOne(t *testing.T) {
	uri := skipIfNoMongoDB(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local client, _ = mongodb.connect({uri = test_uri, database = "test"})
		
		-- Insert test document
		mongodb.insert_one(client, "update_test", {name = "Original", value = 1})
		
		-- Update it
		local result, err = mongodb.update_one(client, "update_test", 
			{name = "Original"},
			{["$set"] = {name = "Updated", value = 2}}
		)
		assert(err == nil, "update_one should not error: " .. tostring(err))
		
		-- Verify
		local doc, _ = mongodb.find_one(client, "update_test", {name = "Updated"})
		assert(doc ~= nil, "updated document should exist")
		assert(doc.value == 2, "value should be updated")
		
		-- Clean up
		mongodb.delete_many(client, "update_test", {})
		mongodb.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// delete tests
// =============================================================================

func TestDeleteOne(t *testing.T) {
	uri := skipIfNoMongoDB(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local client, _ = mongodb.connect({uri = test_uri, database = "test"})
		
		-- Insert test document
		mongodb.insert_one(client, "delete_test", {name = "ToDelete"})
		
		-- Delete it
		local result, err = mongodb.delete_one(client, "delete_test", {name = "ToDelete"})
		assert(err == nil, "delete_one should not error: " .. tostring(err))
		
		-- Verify it's gone
		local doc, _ = mongodb.find_one(client, "delete_test", {name = "ToDelete"})
		assert(doc == nil, "document should be deleted")
		
		mongodb.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// count tests
// =============================================================================

func TestCount(t *testing.T) {
	uri := skipIfNoMongoDB(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local client, _ = mongodb.connect({uri = test_uri, database = "test"})
		
		-- Clear and insert test documents
		mongodb.delete_many(client, "count_test", {})
		mongodb.insert_many(client, "count_test", {
			{category = "A"},
			{category = "A"},
			{category = "B"}
		})
		
		-- Count all
		local count, err = mongodb.count(client, "count_test", {})
		assert(err == nil, "count should not error: " .. tostring(err))
		assert(count == 3, "total count should be 3")
		
		-- Count with filter
		count, err = mongodb.count(client, "count_test", {category = "A"})
		assert(count == 2, "filtered count should be 2")
		
		-- Clean up
		mongodb.delete_many(client, "count_test", {})
		mongodb.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// close tests
// =============================================================================

func TestCloseNil(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local mongodb = require("integrations.mongodb")
		local err = mongodb.close(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
