package redis

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// luaLPush pushes to the left of a list
func luaLPush(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	var values []interface{}
	for i := 3; i <= L.GetTop(); i++ {
		values = append(values, util.LuaToGo(L.Get(i)))
	}

	length, err := w.client.LPush(context.Background(), key, values...)
	if err != nil {
		return util.PushError(L, "lpush failed: %v", err)
	}

	L.Push(lua.LNumber(length))
	L.Push(lua.LNil)
	return 2
}

// luaRPush pushes to the right of a list
func luaRPush(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	var values []interface{}
	for i := 3; i <= L.GetTop(); i++ {
		values = append(values, util.LuaToGo(L.Get(i)))
	}

	length, err := w.client.RPush(context.Background(), key, values...)
	if err != nil {
		return util.PushError(L, "rpush failed: %v", err)
	}

	L.Push(lua.LNumber(length))
	L.Push(lua.LNil)
	return 2
}

// luaLPop pops from the left of a list
func luaLPop(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	val, err := w.client.LPop(context.Background(), key)
	if w.client.IsNilError(err) {
		L.Push(lua.LNil)
		L.Push(lua.LNil)
		return 2
	}
	if err != nil {
		return util.PushError(L, "lpop failed: %v", err)
	}

	L.Push(lua.LString(val))
	L.Push(lua.LNil)
	return 2
}

// luaRPop pops from the right of a list
func luaRPop(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	val, err := w.client.RPop(context.Background(), key)
	if w.client.IsNilError(err) {
		L.Push(lua.LNil)
		L.Push(lua.LNil)
		return 2
	}
	if err != nil {
		return util.PushError(L, "rpop failed: %v", err)
	}

	L.Push(lua.LString(val))
	L.Push(lua.LNil)
	return 2
}

// luaLRange gets a range from a list
func luaLRange(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)
	start := L.CheckInt64(3)
	stop := L.CheckInt64(4)

	values, err := w.client.LRange(context.Background(), key, start, stop)
	if err != nil {
		return util.PushError(L, "lrange failed: %v", err)
	}

	result := L.NewTable()
	for i, v := range values {
		result.RawSetInt(i+1, lua.LString(v))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}
