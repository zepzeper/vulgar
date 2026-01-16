package gitlab

import (
	lua "github.com/yuin/gopher-lua"
	gitlab "github.com/zepzeper/vulgar/internal/services/gitlab"
)

// Usage: local timestamp = gitlab.since_hours(24)
func luaSinceHours(L *lua.LState) int {
	hours := L.CheckInt(1)
	L.Push(lua.LString(gitlab.SinceHours(hours)))
	return 1
}
