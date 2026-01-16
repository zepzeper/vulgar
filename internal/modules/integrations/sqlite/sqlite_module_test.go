package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

// =============================================================================
// open tests
// =============================================================================

func TestOpenMemory(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local sqlite = require("integrations.sqlite")
		local db, err = sqlite.open(":memory:")
		assert(err == nil, "open :memory: should not error: " .. tostring(err))
		assert(db ~= nil, "db should not be nil")
		sqlite.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestOpenFile(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tmpDir := t.TempDir()
	dbFile := filepath.Join(tmpDir, "test.db")
	L.SetGlobal("db_file", lua.LString(dbFile))

	err := L.DoString(`
		local sqlite = require("integrations.sqlite")
		local db, err = sqlite.open(db_file)
		assert(err == nil, "open file should not error: " .. tostring(err))
		assert(db ~= nil, "db should not be nil")
		sqlite.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}
}

func TestOpenInvalidPath(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local sqlite = require("integrations.sqlite")
		local db, err = sqlite.open("/nonexistent/directory/test.db")
		assert(db == nil, "db should be nil for invalid path")
		assert(err ~= nil, "should error for invalid path")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// exec tests
// =============================================================================

func TestExecCreateTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local sqlite = require("integrations.sqlite")
		local db, _ = sqlite.open(":memory:")
		
		local err = sqlite.exec(db, [[
			CREATE TABLE users (
				id INTEGER PRIMARY KEY,
				name TEXT NOT NULL,
				email TEXT UNIQUE
			)
		]])
		assert(err == nil, "exec CREATE TABLE should not error: " .. tostring(err))
		
		sqlite.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExecInsert(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local sqlite = require("integrations.sqlite")
		local db, _ = sqlite.open(":memory:")
		
		sqlite.exec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		
		local err = sqlite.exec(db, "INSERT INTO users (name) VALUES ('John')")
		assert(err == nil, "exec INSERT should not error: " .. tostring(err))
		
		sqlite.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestExecInvalidSQL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local sqlite = require("integrations.sqlite")
		local db, _ = sqlite.open(":memory:")
		
		local err = sqlite.exec(db, "INVALID SQL SYNTAX HERE")
		assert(err ~= nil, "should error for invalid SQL")
		
		sqlite.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// query tests
// =============================================================================

func TestQuerySelect(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local sqlite = require("integrations.sqlite")
		local db, _ = sqlite.open(":memory:")
		
		sqlite.exec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		sqlite.exec(db, "INSERT INTO users (name) VALUES ('Alice')")
		sqlite.exec(db, "INSERT INTO users (name) VALUES ('Bob')")
		
		local rows, err = sqlite.query(db, "SELECT * FROM users")
		assert(err == nil, "query should not error: " .. tostring(err))
		assert(#rows == 2, "should have 2 rows")
		assert(rows[1].name == "Alice", "first row should be Alice")
		assert(rows[2].name == "Bob", "second row should be Bob")
		
		sqlite.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQueryWithParams(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local sqlite = require("integrations.sqlite")
		local db, _ = sqlite.open(":memory:")
		
		sqlite.exec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		sqlite.exec(db, "INSERT INTO users (name) VALUES ('Alice')")
		sqlite.exec(db, "INSERT INTO users (name) VALUES ('Bob')")
		
		local rows, err = sqlite.query(db, "SELECT * FROM users WHERE name = ?", {"Alice"})
		assert(err == nil, "query should not error: " .. tostring(err))
		assert(#rows == 1, "should have 1 row")
		assert(rows[1].name == "Alice", "should find Alice")
		
		sqlite.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQueryEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local sqlite = require("integrations.sqlite")
		local db, _ = sqlite.open(":memory:")
		
		sqlite.exec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		
		local rows, err = sqlite.query(db, "SELECT * FROM users")
		assert(err == nil, "query should not error")
		assert(#rows == 0, "should have 0 rows")
		
		sqlite.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// query_one tests
// =============================================================================

func TestQueryOne(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local sqlite = require("integrations.sqlite")
		local db, _ = sqlite.open(":memory:")
		
		sqlite.exec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		sqlite.exec(db, "INSERT INTO users (name) VALUES ('Alice')")
		
		local row, err = sqlite.query_one(db, "SELECT * FROM users WHERE id = 1")
		assert(err == nil, "query_one should not error: " .. tostring(err))
		assert(row ~= nil, "row should not be nil")
		assert(row.name == "Alice", "name should be Alice")
		
		sqlite.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQueryOneNoResult(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local sqlite = require("integrations.sqlite")
		local db, _ = sqlite.open(":memory:")
		
		sqlite.exec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		
		local row, err = sqlite.query_one(db, "SELECT * FROM users WHERE id = 999")
		assert(row == nil, "row should be nil when no result")
		
		sqlite.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// transaction tests
// =============================================================================

func TestTransaction(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local sqlite = require("integrations.sqlite")
		local db, _ = sqlite.open(":memory:")
		
		sqlite.exec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		
		-- Begin transaction
		local err = sqlite.begin(db)
		assert(err == nil, "begin should not error")
		
		sqlite.exec(db, "INSERT INTO users (name) VALUES ('Test')")
		
		-- Commit
		err = sqlite.commit(db)
		assert(err == nil, "commit should not error")
		
		-- Verify
		local rows, _ = sqlite.query(db, "SELECT * FROM users")
		assert(#rows == 1, "should have 1 row after commit")
		
		sqlite.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestTransactionRollback(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local sqlite = require("integrations.sqlite")
		local db, _ = sqlite.open(":memory:")
		
		sqlite.exec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		
		-- Begin transaction
		sqlite.begin(db)
		sqlite.exec(db, "INSERT INTO users (name) VALUES ('Test')")
		
		-- Rollback
		local err = sqlite.rollback(db)
		assert(err == nil, "rollback should not error")
		
		-- Verify
		local rows, _ = sqlite.query(db, "SELECT * FROM users")
		assert(#rows == 0, "should have 0 rows after rollback")
		
		sqlite.close(db)
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
		local sqlite = require("integrations.sqlite")
		local err = sqlite.close(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
