package redis

import (
	"context"
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// luaGet gets a value by key
func luaGet(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	val, err := w.client.Get(context.Background(), key)
	if w.client.IsNilError(err) {
		L.Push(lua.LNil)
		L.Push(lua.LNil)
		return 2
	}
	if err != nil {
		return util.PushError(L, "get failed: %v", err)
	}

	L.Push(lua.LString(val))
	L.Push(lua.LNil)
	return 2
}

// luaSet sets a value by key
func luaSet(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		L.Push(lua.LString("invalid redis client"))
		return 1
	}

	key := L.CheckString(2)
	value := L.CheckString(3)

	var expiration time.Duration = 0
	if L.GetTop() >= 4 {
		opts := L.OptTable(4, nil)
		if opts != nil {
			if ex := getTableInt(opts, "ex", 0); ex > 0 {
				expiration = time.Duration(ex) * time.Second
			}
			if px := getTableInt(opts, "px", 0); px > 0 {
				expiration = time.Duration(px) * time.Millisecond
			}
		}
	}

	err := w.client.Set(context.Background(), key, value, expiration)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// luaIncr increments a key
func luaIncr(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	val, err := w.client.Incr(context.Background(), key)
	if err != nil {
		return util.PushError(L, "incr failed: %v", err)
	}

	L.Push(lua.LNumber(val))
	L.Push(lua.LNil)
	return 2
}

// luaIncrBy increments a key by a value
func luaIncrBy(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)
	incr := L.CheckInt64(3)

	val, err := w.client.IncrBy(context.Background(), key, incr)
	if err != nil {
		return util.PushError(L, "incrby failed: %v", err)
	}

	L.Push(lua.LNumber(val))
	L.Push(lua.LNil)
	return 2
}
