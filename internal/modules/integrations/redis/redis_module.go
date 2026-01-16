package redis

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.redis"

// luaConnect connects to a Redis server
// Usage: local client, err = redis.connect({host = "localhost", port = 6379, password = "", db = 0})
func luaConnect(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaGet gets a value by key
// Usage: local value, err = redis.get(client, key)
func luaGet(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaSet sets a value by key
// Usage: local err = redis.set(client, key, value, {ex = 3600})
func luaSet(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaDel deletes a key
// Usage: local err = redis.del(client, key)
func luaDel(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaExists checks if a key exists
// Usage: local exists, err = redis.exists(client, key)
func luaExists(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LBool(false))
	L.Push(lua.LString("not implemented"))
	return 2
}

// luaExpire sets expiration on a key
// Usage: local err = redis.expire(client, key, seconds)
func luaExpire(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaHGet gets a hash field
// Usage: local value, err = redis.hget(client, key, field)
func luaHGet(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaHSet sets a hash field
// Usage: local err = redis.hset(client, key, field, value)
func luaHSet(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaHGetAll gets all hash fields
// Usage: local fields, err = redis.hgetall(client, key)
func luaHGetAll(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaLPush pushes to the left of a list
// Usage: local err = redis.lpush(client, key, value)
func luaLPush(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaRPush pushes to the right of a list
// Usage: local err = redis.rpush(client, key, value)
func luaRPush(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaLPop pops from the left of a list
// Usage: local value, err = redis.lpop(client, key)
func luaLPop(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaRPop pops from the right of a list
// Usage: local value, err = redis.rpop(client, key)
func luaRPop(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaPublish publishes a message to a channel
// Usage: local err = redis.publish(client, channel, message)
func luaPublish(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaSubscribe subscribes to a channel
// Usage: local err = redis.subscribe(client, channel, callback)
func luaSubscribe(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaClose closes the Redis connection
// Usage: local err = redis.close(client)
func luaClose(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"connect":   luaConnect,
	"get":       luaGet,
	"set":       luaSet,
	"del":       luaDel,
	"exists":    luaExists,
	"expire":    luaExpire,
	"hget":      luaHGet,
	"hset":      luaHSet,
	"hgetall":   luaHGetAll,
	"lpush":     luaLPush,
	"rpush":     luaRPush,
	"lpop":      luaLPop,
	"rpop":      luaRPop,
	"publish":   luaPublish,
	"subscribe": luaSubscribe,
	"close":     luaClose,
}

// Loader is called when the module is required via require("redis")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
