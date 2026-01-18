package postgres

import (
	"context"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
	postgres "github.com/zepzeper/vulgar/internal/services/postgres"
)

const ModuleName = "integrations.postgres"

// wrapper wraps a Postgres service client
type wrapper struct {
	client *postgres.Client
	tx     *postgres.Tx
}

const (
	dbTypeName = "postgres.db"
	txTypeName = "postgres.tx"
)

// luaConnect connects to PostgreSQL
func luaConnect(L *lua.LState) int {
	var client *postgres.Client
	var err error

	if L.Get(1).Type() == lua.LTString {
		connStr := L.CheckString(1)
		client, err = postgres.NewClient(connStr)
	} else {
		opts := L.CheckTable(1)

		client, err = postgres.NewClientFromOptions(postgres.ConnectOptions{
			Host:     getTableString(opts, "host", "localhost"),
			Port:     getTableInt(opts, "port", 5432),
			User:     getTableString(opts, "user", "postgres"),
			Password: getTableString(opts, "password", ""),
			Database: getTableString(opts, "database", "postgres"),
			SSLMode:  getTableString(opts, "sslmode", "disable"),
		})
	}

	if err != nil {
		return util.PushError(L, "failed to connect: %v", err)
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

// luaQuery executes a query and returns rows
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

	// Convert []map[string]interface{} to Lua table
	result := L.NewTable()
	for i, row := range rows {
		result.RawSetInt(i+1, util.GoToLua(L, row))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaQueryOne executes a query and returns single row
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

// luaExec executes a statement without returning rows
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
	L.SetField(resultTable, "rows_affected", lua.LNumber(res.RowsAffected))

	L.Push(resultTable)
	L.Push(lua.LNil)
	return 2
}

// luaInsert inserts a row
func luaInsert(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid database handle")
	}

	query := L.CheckString(2)
	args := extractArgs(L, 3)

	if strings.Contains(strings.ToUpper(query), "RETURNING") {
		rows, err := w.client.Query(context.Background(), query, args...)
		if err != nil {
			return util.PushError(L, "insert failed: %v", err)
		}
		if len(rows) > 0 {
			L.Push(util.GoToLua(L, rows[0]))
			L.Push(lua.LNil)
			return 2
		}
		L.Push(lua.LNil)
		L.Push(lua.LNil)
		return 2
	}

	res, err := w.client.Exec(context.Background(), query, args...)
	if err != nil {
		return util.PushError(L, "insert failed: %v", err)
	}

	resultTable := L.NewTable()
	L.SetField(resultTable, "rows_affected", lua.LNumber(res.RowsAffected))

	L.Push(resultTable)
	L.Push(lua.LNil)
	return 2
}

func luaUpdate(L *lua.LState) int { return luaExec(L) }
func luaDelete(L *lua.LState) int { return luaExec(L) }

// luaBegin starts a transaction
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

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
