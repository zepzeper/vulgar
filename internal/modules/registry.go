package modules

import "maps"

import lua "github.com/yuin/gopher-lua"

// LuaModule defines the interface for all modules
type LuaModule interface {
	// Loader returns the module table when called via require()
	// This is the standard Lua module loading pattern
	Loader(L *lua.LState) int
}

// PreloadableModule is an optional interface for modules that should be
// available globally without require() (like log)
type PreloadableModule interface {
	LuaModule
	// Open registers the module globally (for critical modules like log)
	Open(L *lua.LState)
}

// registry holds all registered module loaders
var registry = make(map[string]lua.LGFunction)

var preloadRegistry = make(map[string]func(L *lua.LState))

func Register(name string, loader lua.LGFunction) {
	registry[name] = loader
}

func RegisterPreload(name string, opener func(L *lua.LState)) {
	preloadRegistry[name] = opener
}

func GetRegistry() map[string]lua.LGFunction {
	result := make(map[string]lua.LGFunction)
	maps.Copy(result, registry)
	// Same as above
	// for k, v := range registry {
	// 	result[k] = v
	// }
	return result
}

func GetPreloadRegistry() map[string]func(L *lua.LState) {
	result := make(map[string]func(L *lua.LState))
	maps.Copy(result, preloadRegistry)
	return result
}
