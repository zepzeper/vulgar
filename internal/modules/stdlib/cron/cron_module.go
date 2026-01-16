package cron

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
)

var exports = map[string]lua.LGFunction{
	// Core functions
	"schedule": luaSchedule,
	"every":    luaEvery,
	"at":       luaAt,
	"cancel":   luaCancel,
	"start":    luaStart,
	"stop":     luaStop,
	"list":     luaList,
	// Helper functions
	"every_second":          luaEverySecond,
	"every_minute":          luaEveryMinute,
	"every_five_minutes":    luaEveryFiveMinutes,
	"every_fifteen_minutes": luaEveryFifteenMinutes,
	"every_thirty_minutes":  luaEveryThirtyMinutes,
	"every_hour":            luaEveryHour,
	"every_day":             luaEveryDay,
	"every_day_at":          luaEveryDayAt,
	"every_week":            luaEveryWeek,
	"every_weekday":         luaEveryWeekday,
}

func Loader(L *lua.LState) int {
	registerJobType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
