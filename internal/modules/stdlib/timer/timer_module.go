package timer

import (
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const (
	ModuleName       = "stdlib.timer"
	luaTimerTypeName = "timer_handle"
)

type timerHandle struct {
	callback  *lua.LFunction
	L         *lua.LState
	interval  time.Duration
	repeating bool
	done      chan struct{}
	mu        sync.Mutex
	stopped   bool
	queue     *util.EventQueue
}

var timerMethods = map[string]lua.LGFunction{
	"stop":  luaTimerStop,
	"reset": luaReset,
}

func registerTimerType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaTimerTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), timerMethods))
	L.SetField(mt, "__gc", L.NewFunction(timerGC))
}

func checkTimer(L *lua.LState) *timerHandle {
	ud := L.CheckUserData(1)
	if h, ok := ud.Value.(*timerHandle); ok {
		return h
	}
	L.ArgError(1, "timer_handle expected")
	return nil
}

func (h *timerHandle) stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.stopped {
		return
	}
	h.stopped = true
	close(h.done)

	// Notify engine that this source is done
	if h.queue != nil {
		h.queue.RemoveSource()
	}
}

func timerGC(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if h, ok := ud.Value.(*timerHandle); ok {
		// Just stop it, don't remove source again if already stopped
		// We rely on stop()'s idempotency
		h.stop()
	}
	return 0
}

// luaAfter executes a function after a delay
// Usage: local t, err = timer.after(5000, function() print("done") end)
func luaAfter(L *lua.LState) int {
	delayMs := L.CheckNumber(1)
	callback := L.CheckFunction(2)

	if delayMs < 0 {
		return util.PushError(L, "delay cannot be negative")
	}

	queue := util.GetEventQueue(L)
	if queue == nil {
		return util.PushError(L, "engine event queue not initialized")
	}

	delay := time.Duration(delayMs) * time.Millisecond

	h := &timerHandle{
		callback:  callback,
		L:         L,
		interval:  delay,
		repeating: false,
		done:      make(chan struct{}),
		queue:     queue,
	}

	// Register source
	queue.AddSource()

	go func() {
		select {
		case <-time.After(delay):
			h.mu.Lock()
			if !h.stopped {
				// Queue the event
				queue.Queue(h.callback, nil)
				// One-shot timer is done after firing
				h.stopped = true
				close(h.done)
				queue.RemoveSource()
			}
			h.mu.Unlock()
		case <-h.done:
			return
		}
	}()

	ud := util.NewUserData(L, h, luaTimerTypeName)
	return util.PushSuccess(L, ud)
}

// luaEvery executes a function repeatedly at an interval
// Usage: local t, err = timer.every(1000, function() print("tick") end)
func luaEvery(L *lua.LState) int {
	intervalMs := L.CheckNumber(1)
	callback := L.CheckFunction(2)

	if intervalMs <= 0 {
		return util.PushError(L, "interval must be positive")
	}

	queue := util.GetEventQueue(L)
	if queue == nil {
		return util.PushError(L, "engine event queue not initialized")
	}

	interval := time.Duration(intervalMs) * time.Millisecond

	h := &timerHandle{
		callback:  callback,
		L:         L,
		interval:  interval,
		repeating: true,
		done:      make(chan struct{}),
		queue:     queue,
	}

	// Register source
	queue.AddSource()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.mu.Lock()
				if h.stopped {
					h.mu.Unlock()
					return
				}
				queue.Queue(h.callback, nil)
				h.mu.Unlock()
			case <-h.done:
				return
			}
		}
	}()

	ud := util.NewUserData(L, h, luaTimerTypeName)
	return util.PushSuccess(L, ud)
}

// luaCancel cancels a timer
// Usage: timer.stop(t)
func luaCancel(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LNil)
		return 1
	}

	h := util.CheckUserData[*timerHandle](L, 1, luaTimerTypeName)
	if h == nil {
		return util.PushError(L, "invalid timer")
	}

	h.stop()
	L.Push(lua.LNil)
	return 1
}

// luaReset resets a timer with a new delay
// Usage: timer.reset(t, 10000)
func luaReset(L *lua.LState) int {
	h := checkTimer(L)
	newDelayMs := L.CheckNumber(2)

	h.mu.Lock()
	defer h.mu.Unlock()

	// If already stopped, we can't easily restart it because we'd need to re-add to queue source count
	// safely. For simplicity, we only allow resetting active timers, or we'd need to re-increment source.
	if h.stopped {
		// Re-activate
		h.stopped = false
		h.done = make(chan struct{})
		h.queue.AddSource()
	} else {
		// Stop current goroutine
		close(h.done)
		h.done = make(chan struct{})
	}

	h.interval = time.Duration(newDelayMs) * time.Millisecond

	// Restart the timer goroutine
	if h.repeating {
		go func() {
			ticker := time.NewTicker(h.interval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					h.mu.Lock()
					if h.stopped {
						h.mu.Unlock()
						return
					}
					h.queue.Queue(h.callback, nil)
					h.mu.Unlock()
				case <-h.done:
					return
				}
			}
		}()
	} else {
		go func() {
			select {
			case <-time.After(h.interval):
				h.mu.Lock()
				if !h.stopped {
					h.queue.Queue(h.callback, nil)
					h.stopped = true
					close(h.done)
					h.queue.RemoveSource()
				}
				h.mu.Unlock()
			case <-h.done:
				return
			}
		}()
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: timer.sleep(1000)
func luaSleep(L *lua.LState) int {
	amount := L.CheckNumber(1)
	if amount < 0 {
		return 0
	}
	time.Sleep(time.Duration(amount) * time.Millisecond)
	return 0
}

// Usage: local timestamp = timer.now()
func luaNow(L *lua.LState) int {
	L.Push(lua.LNumber(time.Now().UnixMilli()))
	return 1
}

// Usage: local elapsed, result, err = timer.measure(function() return "done" end)
func luaMeasure(L *lua.LState) int {
	action := L.CheckFunction(1)
	start := time.Now()

	L.Push(action)
	err := L.PCall(0, 1, nil)

	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		L.Push(lua.LNumber(elapsed))
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 3
	}

	result := L.Get(-1)
	L.Pop(1)

	L.Push(lua.LNumber(elapsed))
	L.Push(result)
	L.Push(lua.LNil)
	return 3
}

// Usage: t:stop()
func luaTimerStop(L *lua.LState) int {
	h := checkTimer(L)
	h.stop()
	L.Push(lua.LNil)
	return 1
}

var exports = map[string]lua.LGFunction{
	"after":   luaAfter,
	"every":   luaEvery,
	"stop":    luaCancel,
	"reset":   luaReset,
	"sleep":   luaSleep,
	"now":     luaNow,
	"measure": luaMeasure,
}

func Loader(L *lua.LState) int {
	registerTimerType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
