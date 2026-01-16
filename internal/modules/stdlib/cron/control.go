package cron

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// luaStart starts the cron scheduler
// Usage: cron.start()
func luaStart(L *lua.LState) int {
	s := getScheduler()
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		s.c.Start()
		s.started = true
	}

	L.Push(lua.LNil)
	return 1
}

// luaStop stops the cron scheduler
// Usage: cron.stop()
func luaStop(L *lua.LState) int {
	s := getScheduler()
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		s.c.Stop()
		s.started = false
	}

	L.Push(lua.LNil)
	return 1
}

// luaList lists all scheduled jobs
// Usage: local jobs = cron.list()
func luaList(L *lua.LState) int {
	s := getScheduler()
	s.mu.Lock()
	defer s.mu.Unlock()

	tbl := L.NewTable()
	i := 1
	for _, h := range s.jobs {
		jobTbl := L.NewTable()
		jobTbl.RawSetString("expression", lua.LString(h.expr))
		tbl.RawSetInt(i, jobTbl)
		i++
	}

	return util.PushSuccess(L, tbl)
}
