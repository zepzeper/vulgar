package gdrive

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
)

var exports = map[string]lua.LGFunction{
	"configure":     luaConfigure,
	"list_files":    luaListFiles,
	"get_file":      luaGetFile,
	"download":      luaDownload,
	"upload":        luaUpload,
	"delete":        luaDelete,
	"create_folder": luaCreateFolder,
	"move":          luaMove,
	"copy":          luaCopy,
	"share":         luaShare,
	"rename":        luaRename,
	"search":        luaSearch,
	"find_by_name":  luaFindByName,
	"find_by_path":  luaFindByPath,
	"find_folder":   luaFindFolder,
}

// Loader is called when the module is required via require("integrations.gdrive")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
