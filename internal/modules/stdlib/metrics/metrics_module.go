package metrics

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
)

const ModuleName = "stdlib.metrics"

// luaCounter creates or increments a counter
// Usage: metrics.counter("requests_total", 1, {method = "GET", path = "/api"})
func luaCounter(L *lua.LState) int {
	// TODO: implement
	return 0
}

// luaGauge sets a gauge value
// Usage: metrics.gauge("active_connections", 42, {server = "web-1"})
func luaGauge(L *lua.LState) int {
	// TODO: implement
	return 0
}

// luaHistogram records a histogram observation
// Usage: metrics.histogram("request_duration", 0.234, {method = "GET"})
func luaHistogram(L *lua.LState) int {
	// TODO: implement
	return 0
}

// luaSummary records a summary observation
// Usage: metrics.summary("response_size", 1024, {endpoint = "/api/users"})
func luaSummary(L *lua.LState) int {
	// TODO: implement
	return 0
}

// luaTimer times a function execution
// Usage: local result = metrics.timer("operation_duration", function() return do_work() end, {op = "fetch"})
func luaTimer(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNil)
	return 1
}

// luaGet gets the current value of a metric
// Usage: local value = metrics.get("requests_total", {method = "GET"})
func luaGet(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNumber(0))
	return 1
}

// luaExport exports all metrics in Prometheus format
// Usage: local output = metrics.export()
func luaExport(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString(""))
	return 1
}

// luaReset resets all metrics
// Usage: metrics.reset()
func luaReset(L *lua.LState) int {
	// TODO: implement
	return 0
}

var exports = map[string]lua.LGFunction{
	"counter":   luaCounter,
	"gauge":     luaGauge,
	"histogram": luaHistogram,
	"summary":   luaSummary,
	"timer":     luaTimer,
	"get":       luaGet,
	"export":    luaExport,
	"reset":     luaReset,
}

// Loader is called when the module is required
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
