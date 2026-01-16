package postgres

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

// skipIfNoPostgres skips tests if POSTGRES_TEST_URI is not set
func skipIfNoPostgres(t *testing.T) string {
	uri := os.Getenv("POSTGRES_TEST_URI")
	if uri == "" {
		t.Skip("POSTGRES_TEST_URI not set, skipping integration test")
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
		local postgres = require("integrations.postgres")
		local db, err = postgres.connect({})
		-- Should error with missing required config
		assert(db == nil or err ~= nil, "should error with empty config")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectInvalidHost(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local postgres = require("integrations.postgres")
		local db, err = postgres.connect({
			host = "invalid-host-that-does-not-exist",
			port = 5432,
			user = "test",
			password = "test",
			database = "test"
		})
		assert(db == nil, "db should be nil for invalid host")
		assert(err ~= nil, "should return error for invalid host")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectWithURI(t *testing.T) {
	uri := skipIfNoPostgres(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local postgres = require("integrations.postgres")
		local db, err = postgres.connect({uri = test_uri})
		assert(err == nil, "connect should not error: " .. tostring(err))
		assert(db ~= nil, "db should not be nil")
		postgres.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// query tests
// =============================================================================

func TestQueryWithoutConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local postgres = require("integrations.postgres")
		local rows, err = postgres.query(nil, "SELECT 1")
		assert(rows == nil, "rows should be nil without connection")
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQuerySimple(t *testing.T) {
	uri := skipIfNoPostgres(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local postgres = require("integrations.postgres")
		local db, _ = postgres.connect({uri = test_uri})
		local rows, err = postgres.query(db, "SELECT 1 as num")
		assert(err == nil, "query should not error: " .. tostring(err))
		assert(rows ~= nil, "rows should not be nil")
		assert(#rows == 1, "should return 1 row")
		assert(rows[1].num == 1, "value should be 1")
		postgres.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQueryWithParams(t *testing.T) {
	uri := skipIfNoPostgres(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local postgres = require("integrations.postgres")
		local db, _ = postgres.connect({uri = test_uri})
		local rows, err = postgres.query(db, "SELECT $1::int as num", {42})
		assert(err == nil, "query should not error: " .. tostring(err))
		assert(rows[1].num == 42, "parameterized value should be 42")
		postgres.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQueryInvalidSQL(t *testing.T) {
	uri := skipIfNoPostgres(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local postgres = require("integrations.postgres")
		local db, _ = postgres.connect({uri = test_uri})
		local rows, err = postgres.query(db, "INVALID SQL SYNTAX HERE")
		assert(rows == nil, "rows should be nil for invalid SQL")
		assert(err ~= nil, "should return error for invalid SQL")
		postgres.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// query_one tests
// =============================================================================

func TestQueryOne(t *testing.T) {
	uri := skipIfNoPostgres(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local postgres = require("integrations.postgres")
		local db, _ = postgres.connect({uri = test_uri})
		local row, err = postgres.query_one(db, "SELECT 42 as answer")
		assert(err == nil, "query_one should not error: " .. tostring(err))
		assert(row ~= nil, "row should not be nil")
		assert(row.answer == 42, "answer should be 42")
		postgres.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestQueryOneNoRows(t *testing.T) {
	uri := skipIfNoPostgres(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local postgres = require("integrations.postgres")
		local db, _ = postgres.connect({uri = test_uri})
		local row, err = postgres.query_one(db, "SELECT 1 WHERE false")
		-- Should return nil row when no results
		assert(row == nil, "row should be nil when no results")
		postgres.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// exec tests
// =============================================================================

func TestExecWithoutConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local postgres = require("integrations.postgres")
		local result, err = postgres.exec(nil, "SELECT 1")
		assert(result == nil, "result should be nil without connection")
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// transaction tests
// =============================================================================

func TestBeginCommit(t *testing.T) {
	uri := skipIfNoPostgres(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local postgres = require("integrations.postgres")
		local db, _ = postgres.connect({uri = test_uri})
		
		local tx, err = postgres.begin(db)
		assert(err == nil, "begin should not error: " .. tostring(err))
		assert(tx ~= nil, "tx should not be nil")
		
		local err = postgres.commit(tx)
		assert(err == nil, "commit should not error: " .. tostring(err))
		
		postgres.close(db)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestBeginRollback(t *testing.T) {
	uri := skipIfNoPostgres(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local postgres = require("integrations.postgres")
		local db, _ = postgres.connect({uri = test_uri})
		
		local tx, err = postgres.begin(db)
		assert(err == nil, "begin should not error")
		
		local err = postgres.rollback(tx)
		assert(err == nil, "rollback should not error: " .. tostring(err))
		
		postgres.close(db)
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
		local postgres = require("integrations.postgres")
		local err = postgres.close(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestCloseValid(t *testing.T) {
	uri := skipIfNoPostgres(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("test_uri", lua.LString(uri))

	err := L.DoString(`
		local postgres = require("integrations.postgres")
		local db, _ = postgres.connect({uri = test_uri})
		local err = postgres.close(db)
		assert(err == nil, "close should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
