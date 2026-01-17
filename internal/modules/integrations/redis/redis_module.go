package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.redis"

// clientWrapper wraps a Redis client
type clientWrapper struct {
	client *redis.Client
	ctx    context.Context
}

const clientTypeName = "redis.client"

// luaConnect connects to a Redis server
// Usage: local client, err = redis.connect({host = "localhost", port = 6379, password = "", db = 0})
// Or: local client, err = redis.connect("redis://localhost:6379/0")
func luaConnect(L *lua.LState) int {
	var client *redis.Client

	if L.Get(1).Type() == lua.LTString {
		// Connection string
		connStr := L.CheckString(1)
		opt, err := redis.ParseURL(connStr)
		if err != nil {
			return util.PushError(L, "invalid connection string: %v", err)
		}
		client = redis.NewClient(opt)
	} else {
		// Options table
		opts := L.CheckTable(1)

		host := getTableString(opts, "host", "localhost")
		port := getTableInt(opts, "port", 6379)
		password := getTableString(opts, "password", "")
		db := getTableInt(opts, "db", 0)

		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: password,
			DB:       db,
		})
	}

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return util.PushError(L, "failed to connect: %v", err)
	}

	wrapper := &clientWrapper{client: client, ctx: ctx}
	ud := L.NewUserData()
	ud.Value = wrapper
	L.SetMetatable(ud, L.GetTypeMetatable(clientTypeName))

	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

func getClient(L *lua.LState, idx int) *clientWrapper {
	ud := L.CheckUserData(idx)
	if wrapper, ok := ud.Value.(*clientWrapper); ok {
		return wrapper
	}
	return nil
}

// luaGet gets a value by key
// Usage: local value, err = redis.get(client, key)
func luaGet(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	val, err := wrapper.client.Get(wrapper.ctx, key).Result()
	if err == redis.Nil {
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
// Usage: local err = redis.set(client, key, value)
// Usage: local err = redis.set(client, key, value, {ex = 3600})  -- with expiration in seconds
func luaSet(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
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

	err := wrapper.client.Set(wrapper.ctx, key, value, expiration).Err()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// luaDel deletes keys
// Usage: local count, err = redis.del(client, key1, key2, ...)
func luaDel(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	var keys []string
	for i := 2; i <= L.GetTop(); i++ {
		keys = append(keys, L.CheckString(i))
	}

	count, err := wrapper.client.Del(wrapper.ctx, keys...).Result()
	if err != nil {
		return util.PushError(L, "del failed: %v", err)
	}

	L.Push(lua.LNumber(count))
	L.Push(lua.LNil)
	return 2
}

// luaExists checks if keys exist
// Usage: local count, err = redis.exists(client, key1, key2, ...)
func luaExists(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	var keys []string
	for i := 2; i <= L.GetTop(); i++ {
		keys = append(keys, L.CheckString(i))
	}

	count, err := wrapper.client.Exists(wrapper.ctx, keys...).Result()
	if err != nil {
		return util.PushError(L, "exists failed: %v", err)
	}

	L.Push(lua.LNumber(count))
	L.Push(lua.LNil)
	return 2
}

// luaExpire sets expiration on a key
// Usage: local ok, err = redis.expire(client, key, seconds)
func luaExpire(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)
	seconds := L.CheckInt(3)

	ok, err := wrapper.client.Expire(wrapper.ctx, key, time.Duration(seconds)*time.Second).Result()
	if err != nil {
		return util.PushError(L, "expire failed: %v", err)
	}

	L.Push(lua.LBool(ok))
	L.Push(lua.LNil)
	return 2
}

// luaTTL gets the TTL of a key
// Usage: local ttl, err = redis.ttl(client, key)
func luaTTL(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	ttl, err := wrapper.client.TTL(wrapper.ctx, key).Result()
	if err != nil {
		return util.PushError(L, "ttl failed: %v", err)
	}

	L.Push(lua.LNumber(ttl.Seconds()))
	L.Push(lua.LNil)
	return 2
}

// luaKeys gets keys matching a pattern
// Usage: local keys, err = redis.keys(client, "user:*")
func luaKeys(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	pattern := L.CheckString(2)

	keys, err := wrapper.client.Keys(wrapper.ctx, pattern).Result()
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

// luaHGet gets a hash field
// Usage: local value, err = redis.hget(client, key, field)
func luaHGet(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)
	field := L.CheckString(3)

	val, err := wrapper.client.HGet(wrapper.ctx, key, field).Result()
	if err == redis.Nil {
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
// Usage: local err = redis.hset(client, key, field, value)
// Usage: local err = redis.hset(client, key, {field1 = value1, field2 = value2})
func luaHSet(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		L.Push(lua.LString("invalid redis client"))
		return 1
	}

	key := L.CheckString(2)

	var err error
	if L.Get(3).Type() == lua.LTTable {
		// Table of field-value pairs
		tbl := L.CheckTable(3)
		fields := make(map[string]interface{})
		tbl.ForEach(func(k, v lua.LValue) {
			if ks, ok := k.(lua.LString); ok {
				fields[string(ks)] = util.LuaToGo(v)
			}
		})
		err = wrapper.client.HSet(wrapper.ctx, key, fields).Err()
	} else {
		// Single field-value
		field := L.CheckString(3)
		value := L.CheckString(4)
		err = wrapper.client.HSet(wrapper.ctx, key, field, value).Err()
	}

	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// luaHGetAll gets all hash fields
// Usage: local fields, err = redis.hgetall(client, key)
func luaHGetAll(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	fields, err := wrapper.client.HGetAll(wrapper.ctx, key).Result()
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

// luaLPush pushes to the left of a list
// Usage: local length, err = redis.lpush(client, key, value1, value2, ...)
func luaLPush(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	var values []interface{}
	for i := 3; i <= L.GetTop(); i++ {
		values = append(values, util.LuaToGo(L.Get(i)))
	}

	length, err := wrapper.client.LPush(wrapper.ctx, key, values...).Result()
	if err != nil {
		return util.PushError(L, "lpush failed: %v", err)
	}

	L.Push(lua.LNumber(length))
	L.Push(lua.LNil)
	return 2
}

// luaRPush pushes to the right of a list
// Usage: local length, err = redis.rpush(client, key, value1, value2, ...)
func luaRPush(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	var values []interface{}
	for i := 3; i <= L.GetTop(); i++ {
		values = append(values, util.LuaToGo(L.Get(i)))
	}

	length, err := wrapper.client.RPush(wrapper.ctx, key, values...).Result()
	if err != nil {
		return util.PushError(L, "rpush failed: %v", err)
	}

	L.Push(lua.LNumber(length))
	L.Push(lua.LNil)
	return 2
}

// luaLPop pops from the left of a list
// Usage: local value, err = redis.lpop(client, key)
func luaLPop(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	val, err := wrapper.client.LPop(wrapper.ctx, key).Result()
	if err == redis.Nil {
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
// Usage: local value, err = redis.rpop(client, key)
func luaRPop(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	val, err := wrapper.client.RPop(wrapper.ctx, key).Result()
	if err == redis.Nil {
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
// Usage: local values, err = redis.lrange(client, key, 0, -1)
func luaLRange(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)
	start := L.CheckInt64(3)
	stop := L.CheckInt64(4)

	values, err := wrapper.client.LRange(wrapper.ctx, key, start, stop).Result()
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

// luaSAdd adds members to a set
// Usage: local count, err = redis.sadd(client, key, member1, member2, ...)
func luaSAdd(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	var members []interface{}
	for i := 3; i <= L.GetTop(); i++ {
		members = append(members, util.LuaToGo(L.Get(i)))
	}

	count, err := wrapper.client.SAdd(wrapper.ctx, key, members...).Result()
	if err != nil {
		return util.PushError(L, "sadd failed: %v", err)
	}

	L.Push(lua.LNumber(count))
	L.Push(lua.LNil)
	return 2
}

// luaSMembers gets all members of a set
// Usage: local members, err = redis.smembers(client, key)
func luaSMembers(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	members, err := wrapper.client.SMembers(wrapper.ctx, key).Result()
	if err != nil {
		return util.PushError(L, "smembers failed: %v", err)
	}

	result := L.NewTable()
	for i, m := range members {
		result.RawSetInt(i+1, lua.LString(m))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaPublish publishes a message to a channel
// Usage: local count, err = redis.publish(client, channel, message)
func luaPublish(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	channel := L.CheckString(2)
	message := L.CheckString(3)

	count, err := wrapper.client.Publish(wrapper.ctx, channel, message).Result()
	if err != nil {
		return util.PushError(L, "publish failed: %v", err)
	}

	L.Push(lua.LNumber(count))
	L.Push(lua.LNil)
	return 2
}

// luaIncr increments a key
// Usage: local value, err = redis.incr(client, key)
func luaIncr(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	val, err := wrapper.client.Incr(wrapper.ctx, key).Result()
	if err != nil {
		return util.PushError(L, "incr failed: %v", err)
	}

	L.Push(lua.LNumber(val))
	L.Push(lua.LNil)
	return 2
}

// luaIncrBy increments a key by a value
// Usage: local value, err = redis.incrby(client, key, 10)
func luaIncrBy(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)
	incr := L.CheckInt64(3)

	val, err := wrapper.client.IncrBy(wrapper.ctx, key, incr).Result()
	if err != nil {
		return util.PushError(L, "incrby failed: %v", err)
	}

	L.Push(lua.LNumber(val))
	L.Push(lua.LNil)
	return 2
}

// luaClose closes the Redis connection
// Usage: local err = redis.close(client)
func luaClose(L *lua.LState) int {
	wrapper := getClient(L, 1)
	if wrapper == nil {
		L.Push(lua.LString("invalid redis client"))
		return 1
	}

	if err := wrapper.client.Close(); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Helper functions

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

var exports = map[string]lua.LGFunction{
	"connect":  luaConnect,
	"get":      luaGet,
	"set":      luaSet,
	"del":      luaDel,
	"exists":   luaExists,
	"expire":   luaExpire,
	"ttl":      luaTTL,
	"keys":     luaKeys,
	"hget":     luaHGet,
	"hset":     luaHSet,
	"hgetall":  luaHGetAll,
	"lpush":    luaLPush,
	"rpush":    luaRPush,
	"lpop":     luaLPop,
	"rpop":     luaRPop,
	"lrange":   luaLRange,
	"sadd":     luaSAdd,
	"smembers": luaSMembers,
	"publish":  luaPublish,
	"incr":     luaIncr,
	"incrby":   luaIncrBy,
	"close":    luaClose,
}

// Loader is called when the module is required via require("integrations.redis")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
