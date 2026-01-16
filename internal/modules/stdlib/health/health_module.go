package health

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.health"

// luaRegister registers a health check
// Usage: health.register("database", function() return db.ping() == nil end)
func luaRegister(L *lua.LState) int {
	// TODO: implement
	return 0
}

// luaCheck runs all health checks
// Usage: local results, err = health.check()
func luaCheck(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCheckOne runs a single health check
// Usage: local ok, err = health.check_one("database")
func luaCheckOne(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LBool(false))
	L.Push(lua.LString("not implemented"))
	return 2
}

// luaStatus returns the overall health status
// Usage: local status = health.status()
func luaStatus(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("unknown"))
	return 1
}

// luaServe starts an HTTP health check endpoint
// Usage: local server, err = health.serve(":8080", "/health")
func luaServe(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaUnregister removes a health check
// Usage: health.unregister("database")
func luaUnregister(L *lua.LState) int {
	// TODO: implement
	return 0
}

var exports = map[string]lua.LGFunction{
	"register":   luaRegister,
	"check":      luaCheck,
	"check_one":  luaCheckOne,
	"status":     luaStatus,
	"serve":      luaServe,
	"unregister": luaUnregister,
}

// Loader is called when the module is required via require("health")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
