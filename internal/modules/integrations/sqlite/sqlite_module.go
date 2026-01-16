package sqlite

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.sqlite"

// luaOpen opens a SQLite database
// Usage: local db, err = sqlite.open("./data.db")
func luaOpen(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaExec executes a SQL statement without returning rows
// Usage: local err = sqlite.exec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
func luaExec(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaQuery executes a SQL query and returns rows
// Usage: local rows, err = sqlite.query(db, "SELECT * FROM users WHERE id = ?", {1})
func luaQuery(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaQueryOne executes a SQL query and returns a single row
// Usage: local row, err = sqlite.query_one(db, "SELECT * FROM users WHERE id = ?", {1})
func luaQueryOne(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaInsert inserts a row and returns the last insert ID
// Usage: local id, err = sqlite.insert(db, "INSERT INTO users (name) VALUES (?)", {"John"})
func luaInsert(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaUpdate updates rows and returns affected count
// Usage: local count, err = sqlite.update(db, "UPDATE users SET name = ? WHERE id = ?", {"Jane", 1})
func luaUpdate(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaDelete deletes rows and returns affected count
// Usage: local count, err = sqlite.delete(db, "DELETE FROM users WHERE id = ?", {1})
func luaDelete(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaBegin begins a transaction
// Usage: local tx, err = sqlite.begin(db)
func luaBegin(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCommit commits a transaction
// Usage: local err = sqlite.commit(tx)
func luaCommit(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaRollback rolls back a transaction
// Usage: local err = sqlite.rollback(tx)
func luaRollback(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaClose closes the database connection
// Usage: local err = sqlite.close(db)
func luaClose(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
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
	"close":     luaClose,
}

// Loader is called when the module is required via require("sqlite")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
