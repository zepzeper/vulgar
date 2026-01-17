package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	_ "github.com/lib/pq"
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.postgres"

// dbWrapper wraps a PostgreSQL database connection
type dbWrapper struct {
	db *sql.DB
	mu sync.Mutex
}

// txWrapper wraps a PostgreSQL transaction
type txWrapper struct {
	tx *sql.Tx
}

const (
	dbTypeName = "postgres.db"
	txTypeName = "postgres.tx"
)

// luaConnect connects to PostgreSQL
// Usage: local db, err = postgres.connect({host = "localhost", port = 5432, user = "user", password = "pass", database = "mydb"})
// Or: local db, err = postgres.connect("postgres://user:pass@localhost:5432/mydb?sslmode=disable")
func luaConnect(L *lua.LState) int {
	var connStr string

	// Check if first arg is string (connection string) or table (options)
	if L.Get(1).Type() == lua.LTString {
		connStr = L.CheckString(1)
	} else {
		opts := L.CheckTable(1)

		host := getTableString(opts, "host", "localhost")
		port := getTableInt(opts, "port", 5432)
		user := getTableString(opts, "user", "postgres")
		password := getTableString(opts, "password", "")
		database := getTableString(opts, "database", "postgres")
		sslmode := getTableString(opts, "sslmode", "disable")

		connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, database, sslmode)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return util.PushError(L, "failed to open connection: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return util.PushError(L, "failed to connect: %v", err)
	}

	wrapper := &dbWrapper{db: db}
	ud := L.NewUserData()
	ud.Value = wrapper
	L.SetMetatable(ud, L.GetTypeMetatable(dbTypeName))

	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

// getDB extracts the database wrapper from userdata
func getDB(L *lua.LState, idx int) *dbWrapper {
	ud := L.CheckUserData(idx)
	if wrapper, ok := ud.Value.(*dbWrapper); ok {
		return wrapper
	}
	return nil
}

// getTx extracts the transaction wrapper from userdata
func getTx(L *lua.LState, idx int) *txWrapper {
	ud := L.CheckUserData(idx)
	if wrapper, ok := ud.Value.(*txWrapper); ok {
		return wrapper
	}
	return nil
}

// luaQuery executes a query and returns rows
// Usage: local rows, err = postgres.query(db, "SELECT * FROM users WHERE id = $1", {1})
func luaQuery(L *lua.LState) int {
	wrapper := getDB(L, 1)
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

// luaQueryOne executes a query and returns single row
// Usage: local row, err = postgres.query_one(db, "SELECT * FROM users WHERE id = $1", {1})
func luaQueryOne(L *lua.LState) int {
	wrapper := getDB(L, 1)
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

// luaExec executes a statement without returning rows
// Usage: local result, err = postgres.exec(db, "UPDATE users SET name = $1 WHERE id = $2", {"John", 1})
func luaExec(L *lua.LState) int {
	wrapper := getDB(L, 1)
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

	resultTable := L.NewTable()
	if rowsAffected, err := result.RowsAffected(); err == nil {
		L.SetField(resultTable, "rows_affected", lua.LNumber(rowsAffected))
	}

	L.Push(resultTable)
	L.Push(lua.LNil)
	return 2
}

// luaInsert inserts a row using INSERT ... RETURNING
// Usage: local row, err = postgres.insert(db, "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING *", {"John", "john@example.com"})
func luaInsert(L *lua.LState) int {
	wrapper := getDB(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid database handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()

	// If query contains RETURNING, use QueryRow to get the result
	if strings.Contains(strings.ToUpper(query), "RETURNING") {
		rows, err := wrapper.db.Query(query, args...)
		if err != nil {
			return util.PushError(L, "insert failed: %v", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return util.PushError(L, "failed to get columns: %v", err)
		}

		if rows.Next() {
			row, err := scanRow(L, rows, columns)
			if err != nil {
				return util.PushError(L, "failed to scan row: %v", err)
			}
			L.Push(row)
			L.Push(lua.LNil)
			return 2
		}
	}

	// Fallback to Exec
	result, err := wrapper.db.Exec(query, args...)
	if err != nil {
		return util.PushError(L, "insert failed: %v", err)
	}

	resultTable := L.NewTable()
	if rowsAffected, err := result.RowsAffected(); err == nil {
		L.SetField(resultTable, "rows_affected", lua.LNumber(rowsAffected))
	}

	L.Push(resultTable)
	L.Push(lua.LNil)
	return 2
}

// luaUpdate updates rows
// Usage: local count, err = postgres.update(db, "UPDATE users SET name = $1 WHERE id = $2", {"Jane", 1})
func luaUpdate(L *lua.LState) int {
	wrapper := getDB(L, 1)
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

	rowsAffected, _ := result.RowsAffected()
	L.Push(lua.LNumber(rowsAffected))
	L.Push(lua.LNil)
	return 2
}

// luaDelete deletes rows
// Usage: local count, err = postgres.delete(db, "DELETE FROM users WHERE id = $1", {1})
func luaDelete(L *lua.LState) int {
	return luaUpdate(L)
}

// luaBegin starts a transaction
// Usage: local tx, err = postgres.begin(db)
func luaBegin(L *lua.LState) int {
	wrapper := getDB(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid database handle")
	}

	wrapper.mu.Lock()
	tx, err := wrapper.db.Begin()
	wrapper.mu.Unlock()

	if err != nil {
		return util.PushError(L, "failed to begin transaction: %v", err)
	}

	txWrap := &txWrapper{tx: tx}
	ud := L.NewUserData()
	ud.Value = txWrap
	L.SetMetatable(ud, L.GetTypeMetatable(txTypeName))

	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

// luaCommit commits a transaction
// Usage: local err = postgres.commit(tx)
func luaCommit(L *lua.LState) int {
	wrapper := getTx(L, 1)
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
// Usage: local err = postgres.rollback(tx)
func luaRollback(L *lua.LState) int {
	wrapper := getTx(L, 1)
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
// Usage: local result, err = postgres.tx_exec(tx, "INSERT INTO users (name) VALUES ($1)", {"John"})
func luaTxExec(L *lua.LState) int {
	wrapper := getTx(L, 1)
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
	if rowsAffected, err := result.RowsAffected(); err == nil {
		L.SetField(resultTable, "rows_affected", lua.LNumber(rowsAffected))
	}

	L.Push(resultTable)
	L.Push(lua.LNil)
	return 2
}

// luaTxQuery executes a query within a transaction
// Usage: local rows, err = postgres.tx_query(tx, "SELECT * FROM users")
func luaTxQuery(L *lua.LState) int {
	wrapper := getTx(L, 1)
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

// luaClose closes the connection
// Usage: local err = postgres.close(db)
func luaClose(L *lua.LState) int {
	wrapper := getDB(L, 1)
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

// Helper functions

func getTableString(t *lua.LTable, key string, def string) string {
	v := t.RawGetString(key)
	if s, ok := v.(lua.LString); ok {
		return string(s)
	}
	return def
}

func getTableInt(t *lua.LTable, key string, def int) int {
	v := t.RawGetString(key)
	if n, ok := v.(lua.LNumber); ok {
		return int(n)
	}
	return def
}

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
	"connect":   luaConnect,
	"query":     luaQuery,
	"query_one": luaQueryOne,
	"exec":      luaExec,
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

// Loader is called when the module is required via require("integrations.postgres")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
