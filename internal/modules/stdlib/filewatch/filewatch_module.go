package filewatch

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/radovskyb/watcher"
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const (
	ModuleName         = "stdlib.filewatch"
	luaWatcherTypeName = "filewatch_watcher"
	defaultWatchDelay  = 100 * time.Millisecond
)

// watcherHandle wraps a watcher instance with its callback and cleanup
type watcherHandle struct {
	w        *watcher.Watcher
	callback *lua.LFunction
	L        *lua.LState
	mu       sync.Mutex
	closed   bool
	done     chan struct{}
	queue    *util.EventQueue
}

// watcherMethods are methods available on watcher instances (called with : syntax)
var watcherMethods = map[string]lua.LGFunction{
	"close":       luaClose,
	"is_closed":   luaIsClosed,
	"paths":       luaPaths,
	"add_path":    luaAddPath,
	"remove_path": luaRemovePath,
}

func registerWatcherType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaWatcherTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), watcherMethods))
	L.SetField(mt, "__gc", L.NewFunction(watcherGC))
}

func checkWatcher(L *lua.LState) *watcherHandle {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*watcherHandle); ok {
		return v
	}
	L.ArgError(1, "filewatch_watcher expected")
	return nil
}

func (h *watcherHandle) close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}
	h.closed = true

	close(h.done)
	if h.w != nil {
		h.w.Close()
	}

	if h.queue != nil {
		h.queue.RemoveSource()
	}
}

// startEventLoop runs the event loop in a goroutine
func (h *watcherHandle) startEventLoop() {
	go func() {
		for {
			select {
			case event, ok := <-h.w.Event:
				if !ok {
					return
				}
				h.queueEvent(event)
			case err, ok := <-h.w.Error:
				if !ok {
					return
				}
				// Log error but don't call callback
				_ = err
			case <-h.w.Closed:
				return
			case <-h.done:
				return
			}
		}
	}()
}

func (h *watcherHandle) queueEvent(event watcher.Event) {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}

	callback := h.callback
	h.mu.Unlock()

	// Queue task to run on main thread
	h.queue.QueueTask(func(L *lua.LState) {
		eventTbl := L.NewTable()
		eventTbl.RawSetString("path", lua.LString(event.Path))
		eventTbl.RawSetString("type", lua.LString(event.Op.String()))
		eventTbl.RawSetString("old_path", lua.LString(event.OldPath))

		L.Push(callback)
		L.Push(eventTbl)
		if err := L.PCall(1, 0, nil); err != nil {
			_ = err
		}
	})
}

func watcherGC(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if handle, ok := ud.Value.(*watcherHandle); ok {
		handle.close()
	}
	return 0
}

// Usage: local watcher, err = filewatch.watch(path, callback)
func luaWatch(L *lua.LState) int {
	path := L.CheckString(1)
	callback := L.CheckFunction(2)

	if _, err := os.Stat(path); err != nil {
		return util.PushError(L, "path does not exist: %v", err)
	}

	queue := util.GetEventQueue(L)
	if queue == nil {
		return util.PushError(L, "engine event queue not initialized")
	}

	w := watcher.New()
	w.SetMaxEvents(1)

	if err := w.Add(path); err != nil {
		return util.PushError(L, "failed to add path to watcher: %v", err)
	}

	handle := &watcherHandle{
		w:        w,
		callback: callback,
		L:        L,
		done:     make(chan struct{}),
		queue:    queue,
	}

	// Register source
	queue.AddSource()

	go func() {
		if err := w.Start(defaultWatchDelay); err != nil {
			_ = err
		}
	}()

	time.Sleep(50 * time.Millisecond)

	handle.startEventLoop()

	ud := util.NewUserData(L, handle, luaWatcherTypeName)
	return util.PushSuccess(L, ud)
}

// Usage: local err = filewatch.unwatch(watcher)
func luaUnwatch(L *lua.LState) int {
	val := L.Get(1)
	if val == lua.LNil {
		L.Push(lua.LNil)
		return 1
	}

	handle := util.CheckUserData[*watcherHandle](L, 1, luaWatcherTypeName)
	if handle == nil {
		return util.PushError(L, "invalid watcher")
	}

	handle.close()

	L.Push(lua.LNil)
	return 1
}

// Usage: local watcher, err = filewatch.watch_glob(pattern, callback)
func luaWatchGlob(L *lua.LState) int {
	pattern := L.CheckString(1)
	callback := L.CheckFunction(2)

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return util.PushError(L, "invalid glob pattern: %v", err)
	}

	queue := util.GetEventQueue(L)
	if queue == nil {
		return util.PushError(L, "engine event queue not initialized")
	}

	w := watcher.New()
	w.SetMaxEvents(1)

	for _, match := range matches {
		if err := w.Add(match); err != nil {
			w.Close()
			return util.PushError(L, "failed to add path %s to watcher: %v", match, err)
		}
	}

	handle := &watcherHandle{
		w:        w,
		callback: callback,
		L:        L,
		done:     make(chan struct{}),
		queue:    queue,
	}

	// Register source
	queue.AddSource()

	go func() {
		if err := w.Start(defaultWatchDelay); err != nil {
			_ = err
		}
	}()

	time.Sleep(50 * time.Millisecond)

	handle.startEventLoop()

	ud := util.NewUserData(L, handle, luaWatcherTypeName)
	return util.PushSuccess(L, ud)
}

// luaClose closes the watcher and processes any remaining events
// Usage: watcher:close()
func luaClose(L *lua.LState) int {
	handle := checkWatcher(L)
	handle.close()
	L.Push(lua.LNil)
	return 1
}

// luaIsClosed checks if the watcher is closed
// Usage: local closed = watcher:is_closed()
func luaIsClosed(L *lua.LState) int {
	handle := checkWatcher(L)
	handle.mu.Lock()
	closed := handle.closed
	handle.mu.Unlock()
	L.Push(lua.LBool(closed))
	return 1
}

// luaPaths returns the paths being watched
// Usage: local paths = watcher:paths()
func luaPaths(L *lua.LState) int {
	handle := checkWatcher(L)
	handle.mu.Lock()
	defer handle.mu.Unlock()

	if handle.closed || handle.w == nil {
		L.Push(L.NewTable())
		return 1
	}

	paths := L.NewTable()
	index := 1
	for path := range handle.w.WatchedFiles() {
		paths.RawSetInt(index, lua.LString(path))
		index++
	}

	L.Push(paths)
	return 1
}

// luaAddPath adds a path to the watcher
// Usage: local err = watcher:add_path(path)
func luaAddPath(L *lua.LState) int {
	handle := checkWatcher(L)
	path := L.CheckString(2)

	handle.mu.Lock()
	if handle.closed {
		handle.mu.Unlock()
		return util.PushError(L, "watcher closed")
	}

	w := handle.w
	handle.mu.Unlock()

	if w == nil {
		return util.PushError(L, "watcher not initialized")
	}

	if _, err := os.Stat(path); err != nil {
		return util.PushError(L, "path does not exist: %v", err)
	}

	if err := w.Add(path); err != nil {
		return util.PushError(L, "failed to add path to watcher: %v", err)
	}

	L.Push(lua.LNil)
	return 1
}

// luaRemovePath removes a path from the watcher
// Usage: local err = watcher:remove_path(path)
func luaRemovePath(L *lua.LState) int {
	handle := checkWatcher(L)
	path := L.CheckString(2)

	handle.mu.Lock()
	if handle.closed {
		handle.mu.Unlock()
		return util.PushError(L, "watcher is closed")
	}
	w := handle.w
	handle.mu.Unlock()

	if w == nil {
		return util.PushError(L, "watcher not initialized")
	}

	if err := w.Remove(path); err != nil {
		return util.PushError(L, "failed to remove path: %v", err)
	}

	L.Push(lua.LNil)
	return 1
}

var exports = map[string]lua.LGFunction{
	"watch":      luaWatch,
	"unwatch":    luaUnwatch,
	"watch_glob": luaWatchGlob,
}

func Loader(L *lua.LState) int {
	registerWatcherType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
