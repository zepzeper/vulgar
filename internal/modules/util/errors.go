package util

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// PushError pushes a nil result and error message to Lua stack
// Returns 2 (nil result, error string)
// Usage: if err != nil { return util.PushError(L, "operation failed: %v", err) }
func PushError(L *lua.LState, format string, args ...interface{}) int {
	L.Push(lua.LNil)
	L.Push(lua.LString(fmt.Sprintf(format, args...)))
	return 2
}

// PushSuccess pushes a result and nil error to Lua stack
// Returns 2 (result, nil error)
// Usage: return util.PushSuccess(L, result)
func PushSuccess(L *lua.LState, result lua.LValue) int {
	L.Push(result)
	L.Push(lua.LNil)
	return 2
}
