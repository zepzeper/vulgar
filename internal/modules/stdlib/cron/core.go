package cron

import (
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// luaSchedule schedules a job with a cron expression
// Usage: local job, err = cron.schedule("0 * * * * *", function() print("every minute") end)
func luaSchedule(L *lua.LState) int {
	expr := L.CheckString(1)
	callback := L.CheckFunction(2)

	h, err := createJob(L, expr, callback)
	if err != nil {
		return util.PushError(L, "invalid cron expression: %v", err)
	}

	ud := util.NewUserData(L, h, luaJobTypeName)
	return util.PushSuccess(L, ud)
}

// luaEvery schedules a job at a fixed interval
// Usage: local job, err = cron.every("5m", function() print("every 5 minutes") end)
func luaEvery(L *lua.LState) int {
	durationStr := L.CheckString(1)
	callback := L.CheckFunction(2)

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return util.PushError(L, "invalid duration: %v", err)
	}

	expr := fmt.Sprintf("@every %s", duration.String())

	h, err := createJob(L, expr, callback)
	if err != nil {
		return util.PushError(L, "failed to create job: %v", err)
	}

	ud := util.NewUserData(L, h, luaJobTypeName)
	return util.PushSuccess(L, ud)
}

// luaAt schedules a job at a specific time
// Usage: local job, err = cron.at("2024-01-01T00:00:00Z", function() print("happy new year") end)
func luaAt(L *lua.LState) int {
	timeStr := L.CheckString(1)
	callback := L.CheckFunction(2)

	targetTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return util.PushError(L, "invalid time format (use RFC3339): %v", err)
	}

	now := time.Now()
	if targetTime.Before(now) {
		return util.PushError(L, "target time is in the past")
	}

	s := getScheduler()
	queue := util.GetEventQueue(L)
	if queue == nil {
		return util.PushError(L, "engine event queue not initialized")
	}

	h := &jobHandle{
		callback: callback,
		L:        L,
		expr:     timeStr,
		queue:    queue,
	}

	queue.AddSource()

	// Schedule using cron expression for specific time
	expr := fmt.Sprintf("%d %d %d %d %d *",
		targetTime.Second(),
		targetTime.Minute(),
		targetTime.Hour(),
		targetTime.Day(),
		targetTime.Month(),
	)

	id, err := s.c.AddFunc(expr, func() {
		h.mu.Lock()
		if h.stopped {
			h.mu.Unlock()
			return
		}

		queue.Queue(h.callback, nil)
		h.mu.Unlock()

		// One-shot: stop after firing
		h.stop()
	})

	if err != nil {
		queue.RemoveSource()
		return util.PushError(L, "failed to schedule: %v", err)
	}

	h.id = id

	s.mu.Lock()
	s.jobs[id] = h
	if !s.started {
		s.c.Start()
		s.started = true
	}
	s.mu.Unlock()

	ud := util.NewUserData(L, h, luaJobTypeName)
	return util.PushSuccess(L, ud)
}

// luaCancel cancels a scheduled job
// Usage: cron.cancel(job)
func luaCancel(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LNil)
		return 1
	}

	h := util.CheckUserData[*jobHandle](L, 1, luaJobTypeName)
	if h == nil {
		return util.PushError(L, "invalid job")
	}

	h.stop()
	L.Push(lua.LNil)
	return 1
}
