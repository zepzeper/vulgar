package graphql

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.graphql"

// luaNew creates a GraphQL client
// Usage: local client, err = graphql.new({endpoint = "https://api.example.com/graphql", headers = {Authorization = "Bearer ..."}})
func luaNew(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaQuery executes a GraphQL query
// Usage: local result, err = graphql.query(client, "query { users { id name } }", {})
func luaQuery(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaMutate executes a GraphQL mutation
// Usage: local result, err = graphql.mutate(client, "mutation { createUser(name: $name) { id } }", {name = "John"})
func luaMutate(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaSubscribe creates a GraphQL subscription
// Usage: local sub, err = graphql.subscribe(client, "subscription { userCreated { id } }", {}, function(data) print(data) end)
func luaSubscribe(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

var exports = map[string]lua.LGFunction{
	"new":       luaNew,
	"query":     luaQuery,
	"mutate":    luaMutate,
	"subscribe": luaSubscribe,
}

// Loader is called when the module is required via require("graphql")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
