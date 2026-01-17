package sqlite

import (
	"database/sql"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
	_ "modernc.org/sqlite"
)

const ModuleName = "integrations.sqlite"

// dbWrapper wraps a SQLite database connection
type dbWrapper struct {
	db *sql.DB
	mu sync.Mutex
}

// txWrapper wraps a SQLite transaction
type txWrapper struct {
	tx *sql.Tx
}

const (
	dbTypeName = "sqlite.db"
	txTypeName = "sqlite.tx"
)

// luaOpen opens a SQLite database
// Usage: local db, err = sqlite.open("./data.db")
// Options: sqlite.open(":memory:") for in-memory database
func luaOpen(L *lua.LState) int {
	path := L.CheckString(1)

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return util.PushError(L, "failed to open database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return util.PushError(L, "failed to connect to database: %v", err)
	}

	// Enable foreign keys and WAL mode for better performance
	_, _ = db.Exec("PRAGMA foreign_keys = ON")
	_, _ = db.Exec("PRAGMA journal_mode = WAL")

	wrapper := &dbWrapper{db: db}
	ud := L.NewUserData()
	ud.Value = wrapper
	L.SetMetatable(ud, L.GetTypeMetatable(dbTypeName))

	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

// getDB extracts the database wrapper from userdata
func getDB(L *lua.LState, idx int) (*dbWrapper, error) {
	ud := L.CheckUserData(idx)
	if wrapper, ok := ud.Value.(*dbWrapper); ok {
		return wrapper, nil
	}
	return nil, nil
}

// getTx extracts the transaction wrapper from userdata
func getTx(L *lua.LState, idx int) (*txWrapper, error) {
	ud := L.CheckUserData(idx)
	if wrapper, ok := ud.Value.(*txWrapper); ok {
		return wrapper, nil
	}
	return nil, nil
}

// luaExec executes a SQL statement without returning rows
// Usage: local result, err = sqlite.exec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
// Usage: local result, err = sqlite.exec(db, "INSERT INTO users (name) VALUES (?)", {"John"})
func luaExec(L *lua.LState) int {
	wrapper, _ := getDB(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid database handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()

	result, err := wrapper.db.Exec(query, args...)
	if err != nil {
		return util.PushError(L, "exec failed: %v", err)
	}

	// Return result info
	resultTable := L.NewTable()

	if lastID, err := result.LastInsertId(); err == nil {
		L.SetField(resultTable, "last_insert_id", lua.LNumber(lastID))
	}
	if rowsAffected, err := result.RowsAffected(); err == nil {
		L.SetField(resultTable, "rows_affected", lua.LNumber(rowsAffected))
	}

	L.Push(resultTable)
	L.Push(lua.LNil)
	return 2
}

// luaQuery executes a SQL query and returns all rows
// Usage: local rows, err = sqlite.query(db, "SELECT * FROM users WHERE age > ?", {21})
func luaQuery(L *lua.LState) int {
	wrapper, _ := getDB(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid database handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()

	rows, err := wrapper.db.Query(query, args...)
	if err != nil {
		return util.PushError(L, "query failed: %v", err)
	}
	defer rows.Close()

	result, err := rowsToTable(L, rows)
	if err != nil {
		return util.PushError(L, "failed to read rows: %v", err)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaQueryOne executes a SQL query and returns a single row
// Usage: local row, err = sqlite.query_one(db, "SELECT * FROM users WHERE id = ?", {1})
func luaQueryOne(L *lua.LState) int {
	wrapper, _ := getDB(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid database handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()

	rows, err := wrapper.db.Query(query, args...)
	if err != nil {
		return util.PushError(L, "query failed: %v", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return util.PushError(L, "failed to get columns: %v", err)
	}

	if !rows.Next() {
		L.Push(lua.LNil)
		L.Push(lua.LNil)
		return 2
	}

	row, err := scanRow(L, rows, columns)
	if err != nil {
		return util.PushError(L, "failed to scan row: %v", err)
	}

	L.Push(row)
	L.Push(lua.LNil)
	return 2
}

// luaInsert is a convenience wrapper for INSERT that returns the last insert ID
// Usage: local id, err = sqlite.insert(db, "INSERT INTO users (name) VALUES (?)", {"John"})
func luaInsert(L *lua.LState) int {
	wrapper, _ := getDB(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid database handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()

	result, err := wrapper.db.Exec(query, args...)
	if err != nil {
		return util.PushError(L, "insert failed: %v", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return util.PushError(L, "failed to get last insert id: %v", err)
	}

	L.Push(lua.LNumber(lastID))
	L.Push(lua.LNil)
	return 2
}

// luaUpdate is a convenience wrapper for UPDATE that returns rows affected
// Usage: local count, err = sqlite.update(db, "UPDATE users SET name = ? WHERE id = ?", {"Jane", 1})
func luaUpdate(L *lua.LState) int {
	wrapper, _ := getDB(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid database handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()

	result, err := wrapper.db.Exec(query, args...)
	if err != nil {
		return util.PushError(L, "update failed: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return util.PushError(L, "failed to get rows affected: %v", err)
	}

	L.Push(lua.LNumber(rowsAffected))
	L.Push(lua.LNil)
	return 2
}

// luaDelete is a convenience wrapper for DELETE that returns rows affected
// Usage: local count, err = sqlite.delete(db, "DELETE FROM users WHERE id = ?", {1})
func luaDelete(L *lua.LState) int {
	// Same implementation as update - both return rows affected
	return luaUpdate(L)
}

// luaBegin begins a transaction
// Usage: local tx, err = sqlite.begin(db)
func luaBegin(L *lua.LState) int {
	wrapper, _ := getDB(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid database handle")
	}

	wrapper.mu.Lock()
	tx, err := wrapper.db.Begin()
	wrapper.mu.Unlock()

	if err != nil {
		return util.PushError(L, "failed to begin transaction: %v", err)
	}

	txWrapper := &txWrapper{tx: tx}
	ud := L.NewUserData()
	ud.Value = txWrapper
	L.SetMetatable(ud, L.GetTypeMetatable(txTypeName))

	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

// luaCommit commits a transaction
// Usage: local err = sqlite.commit(tx)
func luaCommit(L *lua.LState) int {
	wrapper, _ := getTx(L, 1)
	if wrapper == nil {
		L.Push(lua.LString("invalid transaction handle"))
		return 1
	}

	if err := wrapper.tx.Commit(); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// luaRollback rolls back a transaction
// Usage: local err = sqlite.rollback(tx)
func luaRollback(L *lua.LState) int {
	wrapper, _ := getTx(L, 1)
	if wrapper == nil {
		L.Push(lua.LString("invalid transaction handle"))
		return 1
	}

	if err := wrapper.tx.Rollback(); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// luaTxExec executes within a transaction
// Usage: local result, err = sqlite.tx_exec(tx, "INSERT INTO users (name) VALUES (?)", {"John"})
func luaTxExec(L *lua.LState) int {
	wrapper, _ := getTx(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid transaction handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	result, err := wrapper.tx.Exec(query, args...)
	if err != nil {
		return util.PushError(L, "exec failed: %v", err)
	}

	resultTable := L.NewTable()
	if lastID, err := result.LastInsertId(); err == nil {
		L.SetField(resultTable, "last_insert_id", lua.LNumber(lastID))
	}
	if rowsAffected, err := result.RowsAffected(); err == nil {
		L.SetField(resultTable, "rows_affected", lua.LNumber(rowsAffected))
	}

	L.Push(resultTable)
	L.Push(lua.LNil)
	return 2
}

// luaTxQuery executes a query within a transaction
// Usage: local rows, err = sqlite.tx_query(tx, "SELECT * FROM users")
func luaTxQuery(L *lua.LState) int {
	wrapper, _ := getTx(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid transaction handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	rows, err := wrapper.tx.Query(query, args...)
	if err != nil {
		return util.PushError(L, "query failed: %v", err)
	}
	defer rows.Close()

	result, err := rowsToTable(L, rows)
	if err != nil {
		return util.PushError(L, "failed to read rows: %v", err)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaClose closes the database connection
// Usage: local err = sqlite.close(db)
func luaClose(L *lua.LState) int {
	wrapper, _ := getDB(L, 1)
	if wrapper == nil {
		L.Push(lua.LString("invalid database handle"))
		return 1
	}

	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()

	if err := wrapper.db.Close(); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// extractArgs extracts query arguments from a Lua table
func extractArgs(L *lua.LState, idx int) []interface{} {
	if L.GetTop() < idx {
		return nil
	}

	tbl := L.OptTable(idx, nil)
	if tbl == nil {
		return nil
	}

	var args []interface{}
	tbl.ForEach(func(_, v lua.LValue) {
		args = append(args, util.LuaToGo(v))
	})

	return args
}

// rowsToTable converts SQL rows to a Lua table
func rowsToTable(L *lua.LState, rows *sql.Rows) (*lua.LTable, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := L.NewTable()
	rowNum := 1

	for rows.Next() {
		row, err := scanRow(L, rows, columns)
		if err != nil {
			return nil, err
		}
		result.RawSetInt(rowNum, row)
		rowNum++
	}

	return result, rows.Err()
}

// scanRow scans a single row into a Lua table
func scanRow(L *lua.LState, rows *sql.Rows, columns []string) (*lua.LTable, error) {
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	row := L.NewTable()
	for i, col := range columns {
		L.SetField(row, col, util.GoToLua(L, values[i]))
	}

	return row, nil
}

var exports = map[string]lua.LGFunction{
	"open":      luaOpen,
	"exec":      luaExec,
	"query":     luaQuery,
	"query_one": luaQueryOne,
	"insert":    luaInsert,
	"update":    luaUpdate,
	"delete":    luaDelete,
	"begin":     luaBegin,
	"commit":    luaCommit,
	"rollback":  luaRollback,
	"tx_exec":   luaTxExec,
	"tx_query":  luaTxQuery,
	"close":     luaClose,
}

// Loader is called when the module is required via require("integrations.sqlite")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
