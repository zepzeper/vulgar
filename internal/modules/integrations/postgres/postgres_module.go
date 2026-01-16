package postgres

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.postgres"

// luaConnect connects to PostgreSQL
// Usage: local db, err = postgres.connect({host = "localhost", port = 5432, user = "user", password = "pass", database = "mydb"})
func luaConnect(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaQuery executes a query and returns rows
// Usage: local rows, err = postgres.query(db, "SELECT * FROM users WHERE id = $1", {1})
func luaQuery(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaQueryOne executes a query and returns single row
// Usage: local row, err = postgres.query_one(db, "SELECT * FROM users WHERE id = $1", {1})
func luaQueryOne(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaExec executes a statement without returning rows
// Usage: local result, err = postgres.exec(db, "UPDATE users SET name = $1 WHERE id = $2", {"John", 1})
func luaExec(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaInsert inserts and returns the new row
// Usage: local row, err = postgres.insert(db, "users", {name = "John", email = "john@example.com"}, {returning = "*"})
func luaInsert(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaUpdate updates rows
// Usage: local count, err = postgres.update(db, "users", {name = "Jane"}, {where = "id = $1", args = {1}})
func luaUpdate(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNumber(0))
	L.Push(lua.LString("not implemented"))
	return 2
}

// luaDelete deletes rows
// Usage: local count, err = postgres.delete(db, "users", {where = "id = $1", args = {1}})
func luaDelete(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNumber(0))
	L.Push(lua.LString("not implemented"))
	return 2
}

// luaBegin starts a transaction
// Usage: local tx, err = postgres.begin(db)
func luaBegin(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCommit commits a transaction
// Usage: local err = postgres.commit(tx)
func luaCommit(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaRollback rolls back a transaction
// Usage: local err = postgres.rollback(tx)
func luaRollback(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaClose closes the connection
// Usage: local err = postgres.close(db)
func luaClose(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
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
	"close":     luaClose,
}

// Loader is called when the module is required via require("postgres")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
