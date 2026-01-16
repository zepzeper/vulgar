package event

import (
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.event"

type listener struct {
	callback *lua.LFunction
	once     bool
}

type eventEmitter struct {
	listeners map[string][]*listener
	waiters   map[string][]chan lua.LValue
	mu        sync.RWMutex
}

var globalEmitter *eventEmitter

func getEmitter() *eventEmitter {
	if globalEmitter == nil {
		globalEmitter = &eventEmitter{
			listeners: make(map[string][]*listener),
			waiters:   make(map[string][]chan lua.LValue),
		}
	}
	return globalEmitter
}

func (e *eventEmitter) addListener(eventName string, callback *lua.LFunction, once bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	l := &listener{
		callback: callback,
		once:     once,
	}

	e.listeners[eventName] = append(e.listeners[eventName], l)
}

func (e *eventEmitter) removeListener(eventName string, callback *lua.LFunction) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	listeners, ok := e.listeners[eventName]
	if !ok {
		return false
	}

	for i, l := range listeners {
		if l.callback == callback {
			e.listeners[eventName] = append(listeners[:i], listeners[i+1:]...)
			return true
		}
	}
	return false
}

func (e *eventEmitter) removeAllListeners(eventName string) int {
	e.mu.Lock()
	defer e.mu.Unlock()

	count := len(e.listeners[eventName])
	delete(e.listeners, eventName)
	return count
}

func (e *eventEmitter) getListeners(eventName string) []*listener {
	e.mu.RLock()
	defer e.mu.RUnlock()

	listeners := e.listeners[eventName]
	result := make([]*listener, len(listeners))
	copy(result, listeners)
	return result
}

func (e *eventEmitter) listenerCount(eventName string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.listeners[eventName])
}

func (e *eventEmitter) emit(L *lua.LState, eventName string, data lua.LValue) int {
	listeners := e.getListeners(eventName)

	// Collect once listeners to remove
	var toRemove []*lua.LFunction
	for _, l := range listeners {
		if l.once {
			toRemove = append(toRemove, l.callback)
		}
	}

	// Remove once listeners
	for _, cb := range toRemove {
		e.removeListener(eventName, cb)
	}

	// Call all listeners
	count := 0
	for _, l := range listeners {
		L.Push(l.callback)
		L.Push(data)
		if err := L.PCall(1, 0, nil); err != nil {
			_ = err
		}
		count++
	}

	// Notify waiters
	e.mu.Lock()
	waiters := e.waiters[eventName]
	delete(e.waiters, eventName)
	e.mu.Unlock()

	for _, ch := range waiters {
		select {
		case ch <- data:
		default:
		}
		close(ch)
	}

	return count
}

func (e *eventEmitter) clearAll() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners = make(map[string][]*listener)
}

func (e *eventEmitter) eventNames() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var names []string
	for name, listeners := range e.listeners {
		if len(listeners) > 0 {
			names = append(names, name)
		}
	}
	return names
}

// Usage: local id = event.on("user.created", function(data) print(data.name) end)
func luaOn(L *lua.LState) int {
	eventName := L.CheckString(1)
	callback := L.CheckFunction(2)

	e := getEmitter()
	e.addListener(eventName, callback, false)

	// Return the callback as the "unsubscribe" token
	L.Push(callback)
	return 1
}

// Usage: local id = event.once("startup", function() print("started") end)
func luaOnce(L *lua.LState) int {
	eventName := L.CheckString(1)
	callback := L.CheckFunction(2)

	e := getEmitter()
	e.addListener(eventName, callback, true)

	L.Push(callback)
	return 1
}

// Usage: local count = event.emit("user.created", {name = "John", email = "john@example.com"})
func luaEmit(L *lua.LState) int {
	eventName := L.CheckString(1)
	data := L.Get(2)
	if data == lua.LNil {
		data = L.NewTable()
	}

	e := getEmitter()
	count := e.emit(L, eventName, data)

	L.Push(lua.LNumber(count))
	return 1
}

// Usage: event.off("test", handler) or event.off("test") to remove all
func luaOff(L *lua.LState) int {
	eventName := L.CheckString(1)

	e := getEmitter()

	// If second arg is nil or not provided, remove all listeners
	if L.GetTop() < 2 || L.Get(2) == lua.LNil {
		e.removeAllListeners(eventName)
		return 0
	}

	// Otherwise remove specific listener
	callback := L.CheckFunction(2)
	e.removeListener(eventName, callback)
	return 0
}

// Usage: local count = event.off_all("user.created")
func luaOffAll(L *lua.LState) int {
	eventName := L.CheckString(1)

	e := getEmitter()
	count := e.removeAllListeners(eventName)

	L.Push(lua.LNumber(count))
	return 1
}

// Usage: local count = event.listeners("user.created")
func luaListeners(L *lua.LState) int {
	eventName := L.CheckString(1)

	e := getEmitter()
	count := e.listenerCount(eventName)

	L.Push(lua.LNumber(count))
	return 1
}

// Usage: local count = event.count("user.created")
func luaCount(L *lua.LState) int {
	eventName := L.CheckString(1)

	e := getEmitter()
	count := e.listenerCount(eventName)

	L.Push(lua.LNumber(count))
	return 1
}

// Usage: local events = event.names()
func luaNames(L *lua.LState) int {
	e := getEmitter()
	names := e.eventNames()

	tbl := L.NewTable()
	for i, name := range names {
		tbl.RawSetInt(i+1, lua.LString(name))
	}

	L.Push(tbl)
	return 1
}

// Usage: event.clear()
func luaClear(L *lua.LState) int {
	e := getEmitter()
	e.clearAll()
	return 0
}

// Usage: local data, err = event.wait("user.created", 5000)
func luaWait(L *lua.LState) int {
	_ = L.CheckString(1)
	_ = L.OptNumber(2, 0)

	// Blocking wait not supported in single-threaded Lua
	// Use event.on() with callbacks instead
	return util.PushError(L, "blocking wait not supported; use event.on() with callbacks instead")
}

var exports = map[string]lua.LGFunction{
	"on":        luaOn,
	"once":      luaOnce,
	"emit":      luaEmit,
	"off":       luaOff,
	"off_all":   luaOffAll,
	"listeners": luaListeners,
	"count":     luaCount,
	"names":     luaNames,
	"clear":     luaClear,
	"wait":      luaWait,
}

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
