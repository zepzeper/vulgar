package sqlite

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
	sqlite "github.com/zepzeper/vulgar/internal/services/sqlite"
)

const ModuleName = "integrations.sqlite"

// wrapper wraps a SQLite service client
type wrapper struct {
	client *sqlite.Client
	tx     *sqlite.Tx
}

const (
	dbTypeName = "sqlite.db"
	txTypeName = "sqlite.tx"
)

// luaOpen opens a SQLite database
func luaOpen(L *lua.LState) int {
	path := L.CheckString(1)

	client, err := sqlite.NewClient(path)
	if err != nil {
		return util.PushError(L, "failed to open database: %v", err)
	}

	ud := L.NewUserData()
	ud.Value = &wrapper{client: client}
	L.SetMetatable(ud, L.GetTypeMetatable(dbTypeName))

	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

func getWrapper(L *lua.LState, idx int) *wrapper {
	ud := L.CheckUserData(idx)
	if w, ok := ud.Value.(*wrapper); ok {
		return w
	}
	return nil
}

// luaExec executes a SQL statement
func luaExec(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid database handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	res, err := w.client.Exec(context.Background(), query, args...)
	if err != nil {
		return util.PushError(L, "exec failed: %v", err)
	}

	resultTable := L.NewTable()
	L.SetField(resultTable, "last_insert_id", lua.LNumber(res.LastInsertID))
	L.SetField(resultTable, "rows_affected", lua.LNumber(res.RowsAffected))

	L.Push(resultTable)
	L.Push(lua.LNil)
	return 2
}

// luaQuery executes a SQL query
func luaQuery(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid database handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	rows, err := w.client.Query(context.Background(), query, args...)
	if err != nil {
		return util.PushError(L, "query failed: %v", err)
	}

	result := L.NewTable()
	for i, row := range rows {
		result.RawSetInt(i+1, util.GoToLua(L, row))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaQueryOne executes a SQL query and returns a single row
func luaQueryOne(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid database handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	rows, err := w.client.Query(context.Background(), query, args...)
	if err != nil {
		return util.PushError(L, "query failed: %v", err)
	}

	if len(rows) == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LNil)
		return 2
	}

	L.Push(util.GoToLua(L, rows[0]))
	L.Push(lua.LNil)
	return 2
}

// luaInsert matches luaExec for consistency
func luaInsert(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid database handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	res, err := w.client.Exec(context.Background(), query, args...)
	if err != nil {
		return util.PushError(L, "insert failed: %v", err)
	}

	L.Push(lua.LNumber(res.LastInsertID))
	L.Push(lua.LNil)
	return 2
}

// luaUpdate
func luaUpdate(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid database handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	res, err := w.client.Exec(context.Background(), query, args...)
	if err != nil {
		return util.PushError(L, "update failed: %v", err)
	}

	L.Push(lua.LNumber(res.RowsAffected))
	L.Push(lua.LNil)
	return 2
}

// luaDelete matches luaUpdate
func luaDelete(L *lua.LState) int {
	return luaUpdate(L)
}

// luaBegin begins a transaction
func luaBegin(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid database handle")
	}

	tx, err := w.client.Begin(context.Background())
	if err != nil {
		return util.PushError(L, "failed to begin transaction: %v", err)
	}

	ud := L.NewUserData()
	ud.Value = &wrapper{tx: tx}
	L.SetMetatable(ud, L.GetTypeMetatable(txTypeName))

	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

// luaCommit commits a transaction
func luaCommit(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.tx == nil {
		L.Push(lua.LString("invalid transaction handle"))
		return 1
	}

	if err := w.tx.Commit(); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// luaRollback rolls back a transaction
func luaRollback(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.tx == nil {
		L.Push(lua.LString("invalid transaction handle"))
		return 1
	}

	if err := w.tx.Rollback(); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// luaTxExec executes within a transaction
func luaTxExec(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.tx == nil {
		return util.PushError(L, "invalid transaction handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	res, err := w.tx.Exec(context.Background(), query, args...)
	if err != nil {
		return util.PushError(L, "exec failed: %v", err)
	}

	resultTable := L.NewTable()
	L.SetField(resultTable, "last_insert_id", lua.LNumber(res.LastInsertID))
	L.SetField(resultTable, "rows_affected", lua.LNumber(res.RowsAffected))

	L.Push(resultTable)
	L.Push(lua.LNil)
	return 2
}

// luaTxQuery executes a query within a transaction
func luaTxQuery(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.tx == nil {
		return util.PushError(L, "invalid transaction handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	rows, err := w.tx.Query(context.Background(), query, args...)
	if err != nil {
		return util.PushError(L, "query failed: %v", err)
	}

	result := L.NewTable()
	for i, row := range rows {
		result.RawSetInt(i+1, util.GoToLua(L, row))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaClose closes the database connection
func luaClose(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		L.Push(lua.LString("invalid database handle"))
		return 1
	}

	if err := w.client.Close(); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
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

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
