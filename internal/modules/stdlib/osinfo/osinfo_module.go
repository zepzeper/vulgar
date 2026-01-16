package osinfo

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.osinfo"

// luaHostname returns the system hostname
// Usage: local hostname, err = osinfo.hostname()
func luaHostname(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaPlatform returns OS platform info
// Usage: local info = osinfo.platform()
func luaPlatform(L *lua.LState) int {
	// TODO: implement
	L.Push(L.NewTable())
	return 1
}

// luaCPU returns CPU information
// Usage: local info, err = osinfo.cpu()
func luaCPU(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaMemory returns memory information
// Usage: local info, err = osinfo.memory()
func luaMemory(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaDisk returns disk usage information
// Usage: local info, err = osinfo.disk("/")
func luaDisk(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaNetwork returns network interface information
// Usage: local interfaces, err = osinfo.network()
func luaNetwork(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaUptime returns system uptime
// Usage: local seconds, err = osinfo.uptime()
func luaUptime(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNumber(0))
	L.Push(lua.LString("not implemented"))
	return 2
}

// luaLoadAvg returns system load averages
// Usage: local load1, load5, load15 = osinfo.load_avg()
func luaLoadAvg(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNumber(0))
	L.Push(lua.LNumber(0))
	L.Push(lua.LNumber(0))
	return 3
}

// luaProcesses returns running process information
// Usage: local procs, err = osinfo.processes()
func luaProcesses(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

var exports = map[string]lua.LGFunction{
	"hostname":  luaHostname,
	"platform":  luaPlatform,
	"cpu":       luaCPU,
	"memory":    luaMemory,
	"disk":      luaDisk,
	"network":   luaNetwork,
	"uptime":    luaUptime,
	"load_avg":  luaLoadAvg,
	"processes": luaProcesses,
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
