package cron

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// luaEverySecond schedules a job to run every second
// Usage: local job, err = cron.every_second(function() print("tick") end)
func luaEverySecond(L *lua.LState) int {
	callback := L.CheckFunction(1)
	h, err := createJob(L, "* * * * * *", callback)
	if err != nil {
		return util.PushError(L, "failed to create job: %v", err)
	}
	ud := util.NewUserData(L, h, luaJobTypeName)
	return util.PushSuccess(L, ud)
}

// luaEveryMinute schedules a job to run every minute
// Usage: local job, err = cron.every_minute(function() print("tick") end)
func luaEveryMinute(L *lua.LState) int {
	callback := L.CheckFunction(1)
	h, err := createJob(L, "0 * * * * *", callback)
	if err != nil {
		return util.PushError(L, "failed to create job: %v", err)
	}
	ud := util.NewUserData(L, h, luaJobTypeName)
	return util.PushSuccess(L, ud)
}

// luaEveryFiveMinutes schedules a job to run every 5 minutes
// Usage: local job, err = cron.every_five_minutes(function() print("tick") end)
func luaEveryFiveMinutes(L *lua.LState) int {
	callback := L.CheckFunction(1)
	h, err := createJob(L, "0 */5 * * * *", callback)
	if err != nil {
		return util.PushError(L, "failed to create job: %v", err)
	}
	ud := util.NewUserData(L, h, luaJobTypeName)
	return util.PushSuccess(L, ud)
}

// luaEveryFifteenMinutes schedules a job to run every 15 minutes
// Usage: local job, err = cron.every_fifteen_minutes(function() print("tick") end)
func luaEveryFifteenMinutes(L *lua.LState) int {
	callback := L.CheckFunction(1)
	h, err := createJob(L, "0 */15 * * * *", callback)
	if err != nil {
		return util.PushError(L, "failed to create job: %v", err)
	}
	ud := util.NewUserData(L, h, luaJobTypeName)
	return util.PushSuccess(L, ud)
}

// luaEveryThirtyMinutes schedules a job to run every 30 minutes
// Usage: local job, err = cron.every_thirty_minutes(function() print("tick") end)
func luaEveryThirtyMinutes(L *lua.LState) int {
	callback := L.CheckFunction(1)
	h, err := createJob(L, "0 */30 * * * *", callback)
	if err != nil {
		return util.PushError(L, "failed to create job: %v", err)
	}
	ud := util.NewUserData(L, h, luaJobTypeName)
	return util.PushSuccess(L, ud)
}

// luaEveryHour schedules a job to run every hour
// Usage: local job, err = cron.every_hour(function() print("tick") end)
func luaEveryHour(L *lua.LState) int {
	callback := L.CheckFunction(1)
	h, err := createJob(L, "0 0 * * * *", callback)
	if err != nil {
		return util.PushError(L, "failed to create job: %v", err)
	}
	ud := util.NewUserData(L, h, luaJobTypeName)
	return util.PushSuccess(L, ud)
}

// luaEveryDay schedules a job to run every day at midnight
// Usage: local job, err = cron.every_day(function() print("daily") end)
func luaEveryDay(L *lua.LState) int {
	callback := L.CheckFunction(1)
	h, err := createJob(L, "0 0 0 * * *", callback)
	if err != nil {
		return util.PushError(L, "failed to create job: %v", err)
	}
	ud := util.NewUserData(L, h, luaJobTypeName)
	return util.PushSuccess(L, ud)
}

// luaEveryDayAt schedules a job to run every day at a specific time
// Usage: local job, err = cron.every_day_at("09:00", function() print("9am") end)
func luaEveryDayAt(L *lua.LState) int {
	timeStr := L.CheckString(1)
	callback := L.CheckFunction(2)

	var hour, minute int
	_, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
	if err != nil {
		return util.PushError(L, "invalid time format (use HH:MM): %v", err)
	}

	expr := fmt.Sprintf("0 %d %d * * *", minute, hour)
	h, err := createJob(L, expr, callback)
	if err != nil {
		return util.PushError(L, "failed to create job: %v", err)
	}
	ud := util.NewUserData(L, h, luaJobTypeName)
	return util.PushSuccess(L, ud)
}

// luaEveryWeek schedules a job to run every week on Sunday at midnight
// Usage: local job, err = cron.every_week(function() print("weekly") end)
func luaEveryWeek(L *lua.LState) int {
	callback := L.CheckFunction(1)
	h, err := createJob(L, "0 0 0 * * 0", callback)
	if err != nil {
		return util.PushError(L, "failed to create job: %v", err)
	}
	ud := util.NewUserData(L, h, luaJobTypeName)
	return util.PushSuccess(L, ud)
}

// luaEveryWeekday schedules a job to run every weekday (Mon-Fri) at a specific time
// Usage: local job, err = cron.every_weekday("09:00", function() print("workday") end)
func luaEveryWeekday(L *lua.LState) int {
	timeStr := L.CheckString(1)
	callback := L.CheckFunction(2)

	var hour, minute int
	_, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
	if err != nil {
		return util.PushError(L, "invalid time format (use HH:MM): %v", err)
	}

	expr := fmt.Sprintf("0 %d %d * * 1-5", minute, hour)
	h, err := createJob(L, expr, callback)
	if err != nil {
		return util.PushError(L, "failed to create job: %v", err)
	}
	ud := util.NewUserData(L, h, luaJobTypeName)
	return util.PushSuccess(L, ud)
}
