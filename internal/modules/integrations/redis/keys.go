package redis

import (
	"context"
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// luaDel deletes keys
func luaDel(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	var keys []string
	for i := 2; i <= L.GetTop(); i++ {
		keys = append(keys, L.CheckString(i))
	}

	count, err := w.client.Del(context.Background(), keys...)
	if err != nil {
		return util.PushError(L, "del failed: %v", err)
	}

	L.Push(lua.LNumber(count))
	L.Push(lua.LNil)
	return 2
}

// luaExists checks if keys exist
func luaExists(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	var keys []string
	for i := 2; i <= L.GetTop(); i++ {
		keys = append(keys, L.CheckString(i))
	}

	count, err := w.client.Exists(context.Background(), keys...)
	if err != nil {
		return util.PushError(L, "exists failed: %v", err)
	}

	L.Push(lua.LNumber(count))
	L.Push(lua.LNil)
	return 2
}

// luaExpire sets expiration on a key
func luaExpire(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)
	seconds := L.CheckInt(3)

	ok, err := w.client.Expire(context.Background(), key, time.Duration(seconds)*time.Second)
	if err != nil {
		return util.PushError(L, "expire failed: %v", err)
	}

	L.Push(lua.LBool(ok))
	L.Push(lua.LNil)
	return 2
}

// luaTTL gets the TTL of a key
func luaTTL(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	ttl, err := w.client.TTL(context.Background(), key)
	if err != nil {
		return util.PushError(L, "ttl failed: %v", err)
	}

	L.Push(lua.LNumber(ttl.Seconds()))
	L.Push(lua.LNil)
	return 2
}

// luaKeys gets keys matching a pattern
func luaKeys(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	pattern := L.CheckString(2)

	keys, err := w.client.Keys(context.Background(), pattern)
	if err != nil {
		return util.PushError(L, "keys failed: %v", err)
	}

	result := L.NewTable()
	for i, key := range keys {
		result.RawSetInt(i+1, lua.LString(key))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}
