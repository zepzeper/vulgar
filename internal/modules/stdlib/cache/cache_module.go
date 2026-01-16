package cache

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
)

const ModuleName = "stdlib.cache"

// luaGet gets a value from cache
// Usage: local value, found = cache.get("key")
func luaGet(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNil)
	L.Push(lua.LBool(false))
	return 2
}

// luaSet sets a value in cache with optional TTL
// Usage: cache.set("key", value, {ttl = 3600})
func luaSet(L *lua.LState) int {
	// TODO: implement
	return 0
}

// luaDelete removes a value from cache
// Usage: cache.delete("key")
func luaDelete(L *lua.LState) int {
	// TODO: implement
	return 0
}

// luaExists checks if key exists in cache
// Usage: local exists = cache.exists("key")
func luaExists(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LBool(false))
	return 1
}

// luaClear clears all cache entries
// Usage: cache.clear()
func luaClear(L *lua.LState) int {
	// TODO: implement
	return 0
}

// luaKeys returns all cache keys
// Usage: local keys = cache.keys()
func luaKeys(L *lua.LState) int {
	// TODO: implement
	L.Push(L.NewTable())
	return 1
}

// luaSize returns number of items in cache
// Usage: local count = cache.size()
func luaSize(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNumber(0))
	return 1
}

// luaGetOrSet gets value or sets it if not exists
// Usage: local value = cache.get_or_set("key", function() return compute_value() end, {ttl = 3600})
func luaGetOrSet(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNil)
	return 1
}

// luaIncrement increments a numeric value
// Usage: local new_value = cache.increment("counter", 1)
func luaIncrement(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNumber(0))
	return 1
}

// luaDecrement decrements a numeric value
// Usage: local new_value = cache.decrement("counter", 1)
func luaDecrement(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNumber(0))
	return 1
}

// luaTTL gets remaining TTL for a key
// Usage: local seconds = cache.ttl("key")
func luaTTL(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNumber(-1))
	return 1
}

// luaExpire sets expiration on existing key
// Usage: cache.expire("key", 3600)
func luaExpire(L *lua.LState) int {
	// TODO: implement
	return 0
}

var exports = map[string]lua.LGFunction{
	"get":        luaGet,
	"set":        luaSet,
	"delete":     luaDelete,
	"exists":     luaExists,
	"clear":      luaClear,
	"keys":       luaKeys,
	"size":       luaSize,
	"get_or_set": luaGetOrSet,
	"increment":  luaIncrement,
	"decrement":  luaDecrement,
	"ttl":        luaTTL,
	"expire":     luaExpire,
}

// Loader is called when the module is required via require("cache")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
