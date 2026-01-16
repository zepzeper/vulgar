package cron

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

var jobMethods = map[string]lua.LGFunction{
	"stop": luaJobStop,
}

func registerJobType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaJobTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), jobMethods))
	L.SetField(mt, "__gc", L.NewFunction(jobGC))
}

func checkJob(L *lua.LState) *jobHandle {
	ud := L.CheckUserData(1)
	if h, ok := ud.Value.(*jobHandle); ok {
		return h
	}
	L.ArgError(1, "cron_job expected")
	return nil
}

func (h *jobHandle) stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.stopped {
		return
	}
	h.stopped = true

	s := getScheduler()
	s.mu.Lock()
	s.c.Remove(h.id)
	delete(s.jobs, h.id)
	s.mu.Unlock()

	if h.queue != nil {
		h.queue.RemoveSource()
	}
}

func jobGC(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if h, ok := ud.Value.(*jobHandle); ok {
		h.stop()
	}
	return 0
}

func createJob(L *lua.LState, expr string, callback *lua.LFunction) (*jobHandle, error) {
	s := getScheduler()
	queue := util.GetEventQueue(L)
	if queue == nil {
		return nil, fmt.Errorf("engine event queue not initialized")
	}

	h := &jobHandle{
		callback: callback,
		L:        L,
		expr:     expr,
		queue:    queue,
	}

	// Register source
	queue.AddSource()

	id, err := s.c.AddFunc(expr, func() {
		h.mu.Lock()
		if h.stopped {
			h.mu.Unlock()
			return
		}

		// Queue the event
		queue.Queue(h.callback, nil)
		h.mu.Unlock()
	})

	if err != nil {
		// If failed to add, revert source registration
		queue.RemoveSource()
		return nil, err
	}

	h.id = id

	s.mu.Lock()
	s.jobs[id] = h
	s.mu.Unlock()

	// Ensure scheduler is running
	s.mu.Lock()
	if !s.started {
		s.c.Start()
		s.started = true
	}
	s.mu.Unlock()

	return h, nil
}

// Usage: job:stop()
func luaJobStop(L *lua.LState) int {
	h := checkJob(L)
	h.stop()
	L.Push(lua.LNil)
	return 1
}
