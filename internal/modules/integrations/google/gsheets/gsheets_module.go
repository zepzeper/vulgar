package gsheets

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
)

var exports = map[string]lua.LGFunction{
	"configure":          luaConfigure,
	"get_values":         luaGetValues,
	"set_values":         luaSetValues,
	"append_values":      luaAppendValues,
	"clear_values":       luaClearValues,
	"get_spreadsheet":    luaGetSpreadsheet,
	"create_spreadsheet": luaCreateSpreadsheet,
	"add_sheet":          luaAddSheet,
	"delete_sheet":       luaDeleteSheet,
	"batch_update":       luaBatchUpdate,
	"batch_get_values":   luaBatchGetValues,
	"find_row":           luaFindRow,
	"find_sheet":         luaFindSheet,
	"list_sheets":        luaListSheets,
	"find_column":        luaFindColumn,
}

// Loader is called when the module is required via require("integrations.gsheets")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
