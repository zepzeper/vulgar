package uuid

import (
	"github.com/google/uuid"
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "uuid"

// luaNew generates a new UUID v4
// Usage: local id = uuid.new()
func luaNew(L *lua.LState) int {
	id := uuid.New()
	L.Push(lua.LString(id.String()))
	return 1
}

// luaV4 is an alias for new (generates UUID v4)
// Usage: local id = uuid.v4()
func luaV4(L *lua.LState) int {
	return luaNew(L)
}

// luaParse parses and validates a UUID string
// Usage: local id, err = uuid.parse("...")
func luaParse(L *lua.LState) int {
	str := L.CheckString(1)
	id, err := uuid.Parse(str)
	if err != nil {
		return util.PushError(L, "%v", err)
	}
	return util.PushSuccess(L, lua.LString(id.String()))
}

// luaIsValid checks if a string is a valid UUID
// Usage: local valid = uuid.is_valid("...")
func luaIsValid(L *lua.LState) int {
	str := L.CheckString(1)
	_, err := uuid.Parse(str)
	L.Push(lua.LBool(err == nil))
	return 1
}

// luaVersion returns the version of a UUID
// Usage: local version = uuid.version("...")
func luaVersion(L *lua.LState) int {
	str := L.CheckString(1)
	id, err := uuid.Parse(str)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LNumber(id.Version()))
	L.Push(lua.LNil)
	return 2
}

// exports defines all functions exposed to Lua
var exports = map[string]lua.LGFunction{
	"new":      luaNew,
	"v4":       luaV4,
	"parse":    luaParse,
	"is_valid": luaIsValid,
	"version":  luaVersion,
}

// Loader is called when the module is required via require("uuid")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
