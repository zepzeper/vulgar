package notion

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.notion"

// luaConfigure configures the Notion client
// Usage: local client, err = notion.configure({token = "secret_..."})
func luaConfigure(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaGetPage gets a page by ID
// Usage: local page, err = notion.get_page(client, page_id)
func luaGetPage(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCreatePage creates a new page
// Usage: local page, err = notion.create_page(client, {parent = {database_id = "..."}, properties = {...}})
func luaCreatePage(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaUpdatePage updates a page
// Usage: local page, err = notion.update_page(client, page_id, {properties = {...}})
func luaUpdatePage(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaDeletePage archives/deletes a page
// Usage: local err = notion.delete_page(client, page_id)
func luaDeletePage(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaGetDatabase gets a database by ID
// Usage: local db, err = notion.get_database(client, database_id)
func luaGetDatabase(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaQueryDatabase queries a database
// Usage: local results, err = notion.query_database(client, database_id, {filter = {...}, sorts = {...}})
func luaQueryDatabase(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCreateDatabase creates a new database
// Usage: local db, err = notion.create_database(client, {parent = {page_id = "..."}, title = {...}, properties = {...}})
func luaCreateDatabase(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaAppendBlocks appends blocks to a page
// Usage: local blocks, err = notion.append_blocks(client, page_id, {children = {...}})
func luaAppendBlocks(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaGetBlocks gets child blocks of a block
// Usage: local blocks, err = notion.get_blocks(client, block_id)
func luaGetBlocks(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaSearch searches all pages and databases
// Usage: local results, err = notion.search(client, {query = "search term", filter = {property = "object", value = "page"}})
func luaSearch(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

var exports = map[string]lua.LGFunction{
	"configure":       luaConfigure,
	"get_page":        luaGetPage,
	"create_page":     luaCreatePage,
	"update_page":     luaUpdatePage,
	"delete_page":     luaDeletePage,
	"get_database":    luaGetDatabase,
	"query_database":  luaQueryDatabase,
	"create_database": luaCreateDatabase,
	"append_blocks":   luaAppendBlocks,
	"get_blocks":      luaGetBlocks,
	"search":          luaSearch,
}

// Loader is called when the module is required via require("notion")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
