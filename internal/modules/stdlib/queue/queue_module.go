package queue

import (
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.queue"
const luaQueueTypeName = "queue"

type queueHandle struct {
	name    string
	maxSize int
	items   []lua.LValue
	mu      sync.Mutex
}

var queueMethods = map[string]lua.LGFunction{
	"push":     luaQueuePush,
	"pop":      luaQueuePop,
	"peek":     luaQueuePeek,
	"size":     luaQueueSize,
	"is_empty": luaQueueIsEmpty,
	"clear":    luaQueueClear,
	"to_array": luaQueueToArray,
}

func registerQueueType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaQueueTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), queueMethods))
	L.SetField(mt, "__gc", L.NewFunction(queueGC))
}

func checkQueue(L *lua.LState, idx int) *queueHandle {
	ud := L.CheckUserData(idx)
	if v, ok := ud.Value.(*queueHandle); ok {
		return v
	}
	L.ArgError(idx, "queue expected")
	return nil
}

func queueGC(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if q, ok := ud.Value.(*queueHandle); ok {
		q.mu.Lock()
		q.items = nil
		q.mu.Unlock()
	}
	return 0
}

// Usage: local q, err = queue.new()
// Usage: local q, err = queue.new({max_size = 100, name = "my_queue"})
func luaNew(L *lua.LState) int {
	opts := L.OptTable(1, nil)

	q := &queueHandle{
		name:    "",
		maxSize: 0, // 0 means unlimited
		items:   make([]lua.LValue, 0),
	}

	if opts != nil {
		if v := L.GetField(opts, "max_size"); v != lua.LNil {
			if n, ok := v.(lua.LNumber); ok {
				q.maxSize = int(n)
			}
		}
		if v := L.GetField(opts, "name"); v != lua.LNil {
			q.name = lua.LVAsString(v)
		}
	}

	ud := L.NewUserData()
	ud.Value = q
	L.SetMetatable(ud, L.GetTypeMetatable(luaQueueTypeName))

	return util.PushSuccess(L, ud)
}

// Usage: local err = queue.push(q, item)
func luaPush(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LString("queue is required"))
		return 1
	}

	q := checkQueue(L, 1)
	if q == nil {
		L.Push(lua.LString("invalid queue"))
		return 1
	}

	item := L.Get(2)

	q.mu.Lock()
	defer q.mu.Unlock()

	// Check max size
	if q.maxSize > 0 && len(q.items) >= q.maxSize {
		L.Push(lua.LString("queue is full"))
		return 1
	}

	q.items = append(q.items, item)
	L.Push(lua.LNil)
	return 1
}

// Usage: local item, err = queue.pop(q)
func luaPop(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return util.PushError(L, "queue is required")
	}

	q := checkQueue(L, 1)
	if q == nil {
		return util.PushError(L, "invalid queue")
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LNil)
		return 2
	}

	item := q.items[0]
	q.items = q.items[1:]

	return util.PushSuccess(L, item)
}

// Usage: local item = queue.peek(q)
func luaPeek(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LNil)
		return 1
	}

	q := checkQueue(L, 1)
	if q == nil {
		L.Push(lua.LNil)
		return 1
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		L.Push(lua.LNil)
		return 1
	}

	L.Push(q.items[0])
	return 1
}

// Usage: local size = queue.size(q)
func luaSize(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LNumber(0))
		return 1
	}

	q := checkQueue(L, 1)
	if q == nil {
		L.Push(lua.LNumber(0))
		return 1
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	L.Push(lua.LNumber(len(q.items)))
	return 1
}

// Usage: local empty = queue.is_empty(q)
func luaIsEmpty(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LBool(true))
		return 1
	}

	q := checkQueue(L, 1)
	if q == nil {
		L.Push(lua.LBool(true))
		return 1
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	L.Push(lua.LBool(len(q.items) == 0))
	return 1
}

// Usage: queue.clear(q)
func luaClear(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return 0
	}

	q := checkQueue(L, 1)
	if q == nil {
		return 0
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = make([]lua.LValue, 0)
	return 0
}

// Usage: local arr = queue.to_array(q)
func luaToArray(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(L.NewTable())
		return 1
	}

	q := checkQueue(L, 1)
	if q == nil {
		L.Push(L.NewTable())
		return 1
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	tbl := L.NewTable()
	for i, item := range q.items {
		tbl.RawSetInt(i+1, item)
	}

	L.Push(tbl)
	return 1
}

// Instance methods (for q:method() syntax)

// Usage: q:push(item)
func luaQueuePush(L *lua.LState) int {
	q := checkQueue(L, 1)
	if q == nil {
		L.Push(lua.LString("invalid queue"))
		return 1
	}

	item := L.Get(2)

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.maxSize > 0 && len(q.items) >= q.maxSize {
		L.Push(lua.LString("queue is full"))
		return 1
	}

	q.items = append(q.items, item)
	L.Push(lua.LNil)
	return 1
}

// Usage: q:pop()
func luaQueuePop(L *lua.LState) int {
	q := checkQueue(L, 1)
	if q == nil {
		return util.PushError(L, "invalid queue")
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LNil)
		return 2
	}

	item := q.items[0]
	q.items = q.items[1:]

	return util.PushSuccess(L, item)
}

// Usage: q:peek()
func luaQueuePeek(L *lua.LState) int {
	q := checkQueue(L, 1)
	if q == nil {
		L.Push(lua.LNil)
		return 1
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		L.Push(lua.LNil)
		return 1
	}

	L.Push(q.items[0])
	return 1
}

// Usage: q:size()
func luaQueueSize(L *lua.LState) int {
	q := checkQueue(L, 1)
	if q == nil {
		L.Push(lua.LNumber(0))
		return 1
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	L.Push(lua.LNumber(len(q.items)))
	return 1
}

// Usage: q:is_empty()
func luaQueueIsEmpty(L *lua.LState) int {
	q := checkQueue(L, 1)
	if q == nil {
		L.Push(lua.LBool(true))
		return 1
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	L.Push(lua.LBool(len(q.items) == 0))
	return 1
}

// Usage: q:clear()
func luaQueueClear(L *lua.LState) int {
	q := checkQueue(L, 1)
	if q == nil {
		return 0
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = make([]lua.LValue, 0)
	return 0
}

// Usage: q:to_array()
func luaQueueToArray(L *lua.LState) int {
	q := checkQueue(L, 1)
	if q == nil {
		L.Push(L.NewTable())
		return 1
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	tbl := L.NewTable()
	for i, item := range q.items {
		tbl.RawSetInt(i+1, item)
	}

	L.Push(tbl)
	return 1
}

var exports = map[string]lua.LGFunction{
	"new":      luaNew,
	"push":     luaPush,
	"pop":      luaPop,
	"peek":     luaPeek,
	"size":     luaSize,
	"is_empty": luaIsEmpty,
	"clear":    luaClear,
	"to_array": luaToArray,
}

func Loader(L *lua.LState) int {
	registerQueueType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
