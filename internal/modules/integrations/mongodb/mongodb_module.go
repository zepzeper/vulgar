package mongodb

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.mongodb"

// luaConnect connects to MongoDB
// Usage: local client, err = mongodb.connect({uri = "mongodb://localhost:27017", database = "mydb"})
func luaConnect(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaFindOne finds a single document
// Usage: local doc, err = mongodb.find_one(client, "collection", {_id = "..."})
func luaFindOne(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaFind finds multiple documents
// Usage: local docs, err = mongodb.find(client, "collection", {status = "active"}, {limit = 10, sort = {created = -1}})
func luaFind(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaInsertOne inserts a single document
// Usage: local result, err = mongodb.insert_one(client, "collection", {name = "John", age = 30})
func luaInsertOne(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaInsertMany inserts multiple documents
// Usage: local result, err = mongodb.insert_many(client, "collection", {{name = "John"}, {name = "Jane"}})
func luaInsertMany(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaUpdateOne updates a single document
// Usage: local result, err = mongodb.update_one(client, "collection", {_id = "..."}, {["$set"] = {name = "Updated"}})
func luaUpdateOne(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaUpdateMany updates multiple documents
// Usage: local result, err = mongodb.update_many(client, "collection", {status = "old"}, {["$set"] = {status = "archived"}})
func luaUpdateMany(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaDeleteOne deletes a single document
// Usage: local result, err = mongodb.delete_one(client, "collection", {_id = "..."})
func luaDeleteOne(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaDeleteMany deletes multiple documents
// Usage: local result, err = mongodb.delete_many(client, "collection", {status = "archived"})
func luaDeleteMany(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCount counts documents
// Usage: local count, err = mongodb.count(client, "collection", {status = "active"})
func luaCount(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNumber(0))
	L.Push(lua.LString("not implemented"))
	return 2
}

// luaAggregate runs an aggregation pipeline
// Usage: local results, err = mongodb.aggregate(client, "collection", {{["$match"] = {...}}, {["$group"] = {...}}})
func luaAggregate(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCreateIndex creates an index
// Usage: local err = mongodb.create_index(client, "collection", {email = 1}, {unique = true})
func luaCreateIndex(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaClose closes the connection
// Usage: local err = mongodb.close(client)
func luaClose(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"connect":      luaConnect,
	"find_one":     luaFindOne,
	"find":         luaFind,
	"insert_one":   luaInsertOne,
	"insert_many":  luaInsertMany,
	"update_one":   luaUpdateOne,
	"update_many":  luaUpdateMany,
	"delete_one":   luaDeleteOne,
	"delete_many":  luaDeleteMany,
	"count":        luaCount,
	"aggregate":    luaAggregate,
	"create_index": luaCreateIndex,
	"close":        luaClose,
}

// Loader is called when the module is required via require("mongodb")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
