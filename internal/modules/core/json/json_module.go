package json

import (
	"encoding/json"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "json"

// luaDecode parses a JSON string into a Lua value
// Usage: local data, err = json.decode(json_string)
func luaDecode(L *lua.LState) int {
	jsonString := L.CheckString(1)

	var data interface{}
	if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
		return util.PushError(L, "json decode error: %v", err)
	}

	luaValue := util.GoToLua(L, data)
	return util.PushSuccess(L, luaValue)
}

// luaEncode converts a Lua value to a JSON string
// Usage: local json_string, err = json.encode(lua_table)
func luaEncode(L *lua.LState) int {
	value := L.CheckAny(1)

	// Optional: pretty print
	pretty := false
	if L.GetTop() >= 2 {
		if opts := L.OptTable(2, nil); opts != nil {
			if p := L.GetField(opts, "pretty"); p != lua.LNil {
				pretty = lua.LVAsBool(p)
			}
		}
	}

	goValue := util.LuaToGo(value)

	var result []byte
	var err error
	if pretty {
		result, err = json.MarshalIndent(goValue, "", "  ")
	} else {
		result, err = json.Marshal(goValue)
	}

	if err != nil {
		return util.PushError(L, "json encode error: %v", err)
	}

	return util.PushSuccess(L, lua.LString(string(result)))
}

// exports defines all functions exposed to Lua
var exports = map[string]lua.LGFunction{
	"decode": luaDecode,
	"encode": luaEncode,
}

// Loader is called when the module is required via require("json")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
