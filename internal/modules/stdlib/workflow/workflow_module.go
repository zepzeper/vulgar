package workflow

import (
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.workflow"
const luaWorkflowTypeName = "workflow"

func registerWorkflowType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaWorkflowTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), workflowMethods))
	L.SetField(mt, "__gc", L.NewFunction(workflowGC))
}

func checkWorkflow(L *lua.LState, idx int) *workflowHandle {
	ud := L.CheckUserData(idx)
	if v, ok := ud.Value.(*workflowHandle); ok {
		return v
	}
	L.ArgError(idx, "workflow expected")
	return nil
}

func workflowGC(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if wf, ok := ud.Value.(*workflowHandle); ok {
		wf.mu.Lock()
		wf.nodes = nil
		wf.errorHandler = nil
		wf.context = nil
		wf.mu.Unlock()
	}
	return 0
}

var workflowMethods = map[string]lua.LGFunction{
	"node":      luaWorkflowNode,
	"edge":      luaWorkflowEdge,
	"on_error":  luaWorkflowOnError,
	"run":       luaWorkflowRun,
	"status":    luaWorkflowStatus,
	"cancel":    luaWorkflowCancel,
}

func luaNew(L *lua.LState) int {
	name := L.CheckString(1)
	opts := L.OptTable(2, nil)

	wf := &workflowHandle{
		name:    name,
		nodes:   make(map[string]*workflowNode), // Initialize nodes map
		timeout: 0, // 0 means no timeout
		retries: 0,
		status:  WorkflowStatusPending, // Use new WorkflowStatus type
		context: L.NewTable(),
	}

	if opts != nil {
		if v := L.GetField(opts, "timeout"); v != lua.LNil {
			if n, ok := v.(lua.LNumber); ok {
				wf.timeout = time.Duration(n) * time.Millisecond
			}
		}
		if v := L.GetField(opts, "retries"); v != lua.LNil {
			if n, ok := v.(lua.LNumber); ok {
				wf.retries = int(n)
			}
		}
	}

	ud := L.NewUserData()
	ud.Value = wf
	L.SetMetatable(ud, L.GetTypeMetatable(luaWorkflowTypeName))

	return util.PushSuccess(L, ud)
}

func luaWorkflowNode(L *lua.LState) int {
	return luaNode(L)
}

func luaWorkflowEdge(L *lua.LState) int {
	return luaEdge(L)
}

func luaWorkflowOnError(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return 0
	}

	wf := checkWorkflow(L, 1)
	if wf == nil {
		return 0
	}

	handler := L.CheckFunction(2)

	wf.mu.Lock()
	wf.errorHandler = handler
	wf.mu.Unlock()

	return 0
}

func luaWorkflowStatus(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(L.NewTable())
		return 1
	}

	wf := checkWorkflow(L, 1)
	if wf == nil {
		L.Push(L.NewTable())
		return 1
	}

	wf.mu.Lock()
	status := wf.status
	nodeCount := len(wf.nodes)
	wf.mu.Unlock()

	tbl := L.NewTable()
	tbl.RawSetString("status", lua.LString(status))
	tbl.RawSetString("name", lua.LString(wf.name))
	tbl.RawSetString("nodes", lua.LNumber(nodeCount))

	L.Push(tbl)
	return 1
}

func luaWorkflowCancel(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LNil)
		return 1
	}

	wf := checkWorkflow(L, 1)
	if wf == nil {
		L.Push(lua.LNil)
		return 1
	}

	wf.mu.Lock()
	wf.cancelled = true
	if wf.status == WorkflowStatusPending {
		wf.status = WorkflowStatusCancelled
	}
	wf.mu.Unlock()

	L.Push(lua.LNil)
	return 1
}

var exports = map[string]lua.LGFunction{
	"new":       luaNew,
	"node":      luaNode,
	"edge":      luaEdge,
	"on_error":  luaOnError,
	"run":       luaRun,
	"status":    luaStatus,
	"cancel":    luaCancel,
}

func luaOnError(L *lua.LState) int {
	return luaWorkflowOnError(L)
}

func luaRun(L *lua.LState) int {
	return luaWorkflowRun(L)
}

func luaStatus(L *lua.LState) int {
	return luaWorkflowStatus(L)
}

func luaCancel(L *lua.LState) int {
	return luaWorkflowCancel(L)
}

func Loader(L *lua.LState) int {
	registerWorkflowType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
