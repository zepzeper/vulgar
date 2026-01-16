package env

import (
	"os"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
)

const ModuleName = "env"

// luaGet retrieves an environment variable
// Usage: local value = env.get("VAR_NAME") or env.get("VAR_NAME", "default")
func luaGet(L *lua.LState) int {
	name := L.CheckString(1)
	defaultValue := L.OptString(2, "")

	value, exists := os.LookupEnv(name)
	if !exists {
		if L.GetTop() >= 2 {
			L.Push(lua.LString(defaultValue))
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}

	L.Push(lua.LString(value))
	return 1
}

// luaSet sets an environment variable
// Usage: env.set("VAR_NAME", "value")
func luaSet(L *lua.LState) int {
	name := L.CheckString(1)
	value := L.CheckString(2)

	if err := os.Setenv(name, value); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// luaExists checks if an environment variable exists
// Usage: local exists = env.exists("VAR_NAME")
func luaExists(L *lua.LState) int {
	name := L.CheckString(1)

	_, exists := os.LookupEnv(name)
	L.Push(lua.LBool(exists))
	return 1
}

func luaUnset(L *lua.LState) int {
	name := L.CheckString(1)

	if err := os.Unsetenv(name); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

func luaAll(L *lua.LState) int {
	tbl := L.NewTable()

	for _, e := range os.Environ() {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				key := e[:i]
				value := e[i+1:]
				tbl.RawSetString(key, lua.LString(value))
				break
			}
		}
	}

	L.Push(tbl)
	return 1
}

var exports = map[string]lua.LGFunction{
	"get":    luaGet,
	"set":    luaSet,
	"exists": luaExists,
	"unset":  luaUnset,
	"all":    luaAll,
}

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
