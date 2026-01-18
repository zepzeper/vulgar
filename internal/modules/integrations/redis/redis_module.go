package redis

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
	redis "github.com/zepzeper/vulgar/internal/services/redis"
)

const ModuleName = "integrations.redis"

// wrapper wraps a Redis service client
type wrapper struct {
	client *redis.Client
}

const clientTypeName = "redis.client"

// luaConnect connects to a Redis server
func luaConnect(L *lua.LState) int {
	var client *redis.Client
	var err error

	if L.Get(1).Type() == lua.LTString {
		connStr := L.CheckString(1)
		client, err = redis.NewClient(connStr)
	} else {
		opts := L.CheckTable(1)

		client, err = redis.NewClientFromOptions(redis.ConnectOptions{
			Host:     getTableString(opts, "host", "localhost"),
			Port:     getTableInt(opts, "port", 6379),
			Password: getTableString(opts, "password", ""),
			DB:       getTableInt(opts, "db", 0),
		})
	}

	if err != nil {
		return util.PushError(L, "failed to connect: %v", err)
	}

	ud := L.NewUserData()
	ud.Value = &wrapper{client: client}
	L.SetMetatable(ud, L.GetTypeMetatable(clientTypeName))

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

// luaClose closes the Redis connection
func luaClose(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		L.Push(lua.LString("invalid redis client"))
		return 1
	}

	if err := w.client.Close(); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Helpers

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

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
