package yaml

import (
	"os"

	"github.com/go-yaml/yaml"
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.yaml"

// luaDecode parses a YAML string into a Lua value
// Usage: local data, err = yaml.decode(yaml_string)
func luaDecode(L *lua.LState) int {
	yamlString := L.CheckString(1)

	var data interface{}
	if err := yaml.Unmarshal([]byte(yamlString), &data); err != nil {
		return util.PushError(L, "yaml decode error: %v", err)
	}

	luaValue := util.GoToLua(L, data)
	return util.PushSuccess(L, luaValue)
}

// luaEncode converts a Lua value to a YAML string
// Usage: local yaml_string, err = yaml.encode(lua_table)
func luaEncode(L *lua.LState) int {
	yamlTable := L.CheckTable(1)

	yamlGo := util.LuaToGo(yamlTable)

	data, err := yaml.Marshal(yamlGo)
	if err != nil {
		return util.PushError(L, "yaml encode error: %v", err)
	}

	// Return YAML as a Lua string
	luaValue := lua.LString(data)

	return util.PushSuccess(L, luaValue)
}

// luaDecodeFile reads and parses a YAML file
// Usage: local data, err = yaml.decode_file(path)
func luaDecodeFile(L *lua.LState) int {
	filePath := L.CheckString(1)

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return util.PushError(L, "file read error: %v", err)
	}

	// Parse YAML
	var data interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return util.PushError(L, "yaml decode error: %v", err)
	}

	// Convert to Lua value
	luaValue := util.GoToLua(L, data)
	return util.PushSuccess(L, luaValue)
}

// luaEncodeFile encodes data and writes to a YAML file
// Usage: local err = yaml.encode_file(path, data)
func luaEncodeFile(L *lua.LState) int {
	filePath := L.CheckString(1)
	dataTable := L.CheckTable(2)

	// Convert Lua table to Go value
	dataGo := util.LuaToGo(dataTable)

	// Marshal to YAML
	yamlData, err := yaml.Marshal(dataGo)
	if err != nil {
		return util.PushError(L, "yaml encode error: %v", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
		return util.PushError(L, "file write error: %v", err)
	}

	return util.PushSuccess(L, lua.LNil)
}

var exports = map[string]lua.LGFunction{
	"decode":      luaDecode,
	"encode":      luaEncode,
	"decode_file": luaDecodeFile,
	"encode_file": luaEncodeFile,
}

// Loader is called when the module is required via require("yaml")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
