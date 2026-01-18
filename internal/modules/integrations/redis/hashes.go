package redis

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// luaHGet gets a hash field
func luaHGet(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)
	field := L.CheckString(3)

	val, err := w.client.HGet(context.Background(), key, field)
	if w.client.IsNilError(err) {
		L.Push(lua.LNil)
		L.Push(lua.LNil)
		return 2
	}
	if err != nil {
		return util.PushError(L, "hget failed: %v", err)
	}

	L.Push(lua.LString(val))
	L.Push(lua.LNil)
	return 2
}

// luaHSet sets hash fields
func luaHSet(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		L.Push(lua.LString("invalid redis client"))
		return 1
	}

	key := L.CheckString(2)

	var err error
	if L.Get(3).Type() == lua.LTTable {
		// Table of field-value pairs
		tbl := L.CheckTable(3)
		// Convert Loop
		// Note: The service expects values...interface{}.
		// go-redis can take map[string]interface{}.
		// But our service wrapper defines HSet(key, values...)
		// We should flatten the map to pairs or support map in service?
		// Checking service: return c.client.HSet(ctx, key, values...).Result()
		// go-redis HSet accepts map or pairs. So passing items is fine.

		var args []interface{}
		tbl.ForEach(func(k, v lua.LValue) {
			args = append(args, util.LuaToGo(k), util.LuaToGo(v))
		})
		_, err = w.client.HSet(context.Background(), key, args...)
	} else {
		// Single field-value
		field := L.CheckString(3)
		value := L.CheckString(4)
		_, err = w.client.HSet(context.Background(), key, field, value)
	}

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// luaHGetAll gets all hash fields
func luaHGetAll(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	fields, err := w.client.HGetAll(context.Background(), key)
	if err != nil {
		return util.PushError(L, "hgetall failed: %v", err)
	}

	result := L.NewTable()
	for k, v := range fields {
		L.SetField(result, k, lua.LString(v))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}
