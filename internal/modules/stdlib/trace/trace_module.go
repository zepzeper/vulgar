package trace

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.trace"

// luaStart starts a new trace span
// Usage: local span, err = trace.start("operation_name", {parent = parent_span})
func luaStart(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaEnd ends a trace span
// Usage: trace.end(span)
func luaEnd(L *lua.LState) int {
	// TODO: implement
	return 0
}

// luaAddEvent adds an event to a span
// Usage: trace.add_event(span, "cache_hit", {key = "user:123"})
func luaAddEvent(L *lua.LState) int {
	// TODO: implement
	return 0
}

// luaSetAttribute sets an attribute on a span
// Usage: trace.set_attribute(span, "user.id", "123")
func luaSetAttribute(L *lua.LState) int {
	// TODO: implement
	return 0
}

// luaSetStatus sets the status of a span
// Usage: trace.set_status(span, "error", "something went wrong")
func luaSetStatus(L *lua.LState) int {
	// TODO: implement
	return 0
}

// luaGetContext gets the trace context for propagation
// Usage: local ctx = trace.get_context(span)
func luaGetContext(L *lua.LState) int {
	// TODO: implement
	L.Push(L.NewTable())
	return 1
}

// luaFromContext creates a span from propagated context
// Usage: local span, err = trace.from_context("operation", context_headers)
func luaFromContext(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaWrap wraps a function with tracing
// Usage: local result = trace.wrap("operation", function() return do_work() end)
func luaWrap(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNil)
	return 1
}

var exports = map[string]lua.LGFunction{
	"start":         luaStart,
	"end":           luaEnd,
	"add_event":     luaAddEvent,
	"set_attribute": luaSetAttribute,
	"set_status":    luaSetStatus,
	"get_context":   luaGetContext,
	"from_context":  luaFromContext,
	"wrap":          luaWrap,
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
