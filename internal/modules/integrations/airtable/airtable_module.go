package airtable

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.airtable"

// luaConfigure configures the Airtable client
// Usage: local client, err = airtable.configure({api_key = "pat...", base_id = "app..."})
func luaConfigure(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaListRecords lists records from a table
// Usage: local records, err = airtable.list_records(client, "Table Name", {view = "Grid view", max_records = 100})
func luaListRecords(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaGetRecord gets a single record
// Usage: local record, err = airtable.get_record(client, "Table Name", record_id)
func luaGetRecord(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCreateRecord creates a new record
// Usage: local record, err = airtable.create_record(client, "Table Name", {Name = "John", Email = "john@example.com"})
func luaCreateRecord(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCreateRecords creates multiple records
// Usage: local records, err = airtable.create_records(client, "Table Name", {{Name = "John"}, {Name = "Jane"}})
func luaCreateRecords(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaUpdateRecord updates a record
// Usage: local record, err = airtable.update_record(client, "Table Name", record_id, {Name = "Updated"})
func luaUpdateRecord(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaUpdateRecords updates multiple records
// Usage: local records, err = airtable.update_records(client, "Table Name", {{id = "rec1", fields = {...}}, ...})
func luaUpdateRecords(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaDeleteRecord deletes a record
// Usage: local err = airtable.delete_record(client, "Table Name", record_id)
func luaDeleteRecord(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaDeleteRecords deletes multiple records
// Usage: local err = airtable.delete_records(client, "Table Name", {record_id1, record_id2})
func luaDeleteRecords(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"configure":      luaConfigure,
	"list_records":   luaListRecords,
	"get_record":     luaGetRecord,
	"create_record":  luaCreateRecord,
	"create_records": luaCreateRecords,
	"update_record":  luaUpdateRecord,
	"update_records": luaUpdateRecords,
	"delete_record":  luaDeleteRecord,
	"delete_records": luaDeleteRecords,
}

// Loader is called when the module is required via require("airtable")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
