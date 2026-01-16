package time

import (
	gotime "time"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "time"

// luaNow returns the current timestamp
// Usage: local ts = time.now() -- returns Unix timestamp
func luaNow(L *lua.LState) int {
	L.Push(lua.LNumber(gotime.Now().Unix()))
	return 1
}

// luaNowMs returns the current timestamp in milliseconds
// Usage: local ts = time.now_ms()
func luaNowMs(L *lua.LState) int {
	L.Push(lua.LNumber(gotime.Now().UnixMilli()))
	return 1
}

// luaFormat formats a timestamp with the given layout
// Usage: local str = time.format(timestamp, "2006-01-02 15:04:05")
// If no timestamp provided, uses current time
func luaFormat(L *lua.LState) int {
	var t gotime.Time

	if L.GetTop() >= 1 && L.Get(1) != lua.LNil {
		ts := L.CheckNumber(1)
		t = gotime.Unix(int64(ts), 0)
	} else {
		t = gotime.Now()
	}

	layout := L.OptString(2, "2006-01-02 15:04:05")
	L.Push(lua.LString(t.Format(layout)))
	return 1
}

// luaParse parses a time string with the given layout
// Usage: local ts, err = time.parse("2024-01-15", "2006-01-02")
func luaParse(L *lua.LState) int {
	str := L.CheckString(1)
	layout := L.OptString(2, "2006-01-02 15:04:05")

	t, err := gotime.Parse(layout, str)
	if err != nil {
		return util.PushError(L, "failed to parse time: %v", err)
	}

	return util.PushSuccess(L, lua.LNumber(t.Unix()))
}

// luaSleep pauses execution for the specified duration
// Usage: time.sleep(1.5) -- sleeps for 1.5 seconds
func luaSleep(L *lua.LState) int {
	seconds := L.CheckNumber(1)
	gotime.Sleep(gotime.Duration(seconds * lua.LNumber(gotime.Second)))
	return 0
}

// luaDate returns date components for a timestamp
// Usage: local date = time.date(timestamp)
func luaDate(L *lua.LState) int {
	var t gotime.Time

	if L.GetTop() >= 1 && L.Get(1) != lua.LNil {
		ts := L.CheckNumber(1)
		t = gotime.Unix(int64(ts), 0)
	} else {
		t = gotime.Now()
	}

	tbl := L.NewTable()
	tbl.RawSetString("year", lua.LNumber(t.Year()))
	tbl.RawSetString("month", lua.LNumber(t.Month()))
	tbl.RawSetString("day", lua.LNumber(t.Day()))
	tbl.RawSetString("hour", lua.LNumber(t.Hour()))
	tbl.RawSetString("minute", lua.LNumber(t.Minute()))
	tbl.RawSetString("second", lua.LNumber(t.Second()))
	tbl.RawSetString("weekday", lua.LNumber(t.Weekday()))
	tbl.RawSetString("yearday", lua.LNumber(t.YearDay()))

	L.Push(tbl)
	return 1
}

// luaAdd adds duration to a timestamp
// Usage: local new_ts = time.add(timestamp, 3600) -- add 1 hour
func luaAdd(L *lua.LState) int {
	ts := L.CheckNumber(1)
	seconds := L.CheckNumber(2)

	t := gotime.Unix(int64(ts), 0)
	t = t.Add(gotime.Duration(seconds) * gotime.Second)

	L.Push(lua.LNumber(t.Unix()))
	return 1
}

// luaSub returns the difference between two timestamps in seconds
// Usage: local diff = time.sub(ts1, ts2)
func luaSub(L *lua.LState) int {
	ts1 := L.CheckNumber(1)
	ts2 := L.CheckNumber(2)

	t1 := gotime.Unix(int64(ts1), 0)
	t2 := gotime.Unix(int64(ts2), 0)

	L.Push(lua.LNumber(t1.Sub(t2).Seconds()))
	return 1
}

// luaUtc converts a timestamp to UTC
// Usage: local utc_ts = time.utc(timestamp)
func luaUtc(L *lua.LState) int {
	ts := L.CheckNumber(1)
	t := gotime.Unix(int64(ts), 0).UTC()
	L.Push(lua.LNumber(t.Unix()))
	return 1
}

// exports defines all functions exposed to Lua
var exports = map[string]lua.LGFunction{
	"now":    luaNow,
	"now_ms": luaNowMs,
	"format": luaFormat,
	"parse":  luaParse,
	"sleep":  luaSleep,
	"date":   luaDate,
	"add":    luaAdd,
	"sub":    luaSub,
	"utc":    luaUtc,
}

// Loader is called when the module is required via require("time")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
