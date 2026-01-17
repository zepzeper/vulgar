package workflow

import (
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type NodeStatus string

const (
	NodeStatusPending   NodeStatus = "pending"
	NodeStatusRunning   NodeStatus = "running"
	NodeStatusCompleted NodeStatus = "completed"
	NodeStatusFailed    NodeStatus = "failed"
	NodeStatusSkipped   NodeStatus = "skipped"
)

type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

type workflowNode struct {
	name         string
	fn           *lua.LFunction
	dependencies []string // Names of nodes that must complete before this one
	outputs      []string // Names of nodes that depend on this one (computed from graph)
	status       NodeStatus
	result       lua.LValue // Result from execution
	mu           sync.Mutex
}

type workflowHandle struct {
	name         string
	nodes        map[string]*workflowNode // Graph: node name -> node
	timeout      time.Duration
	retries      int
	status       WorkflowStatus
	errorHandler *lua.LFunction
	context      *lua.LTable // Shared context (merged from all nodes)
	output       lua.LValue  // Final output (not a pointer - lua.LValue is an interface)
	mu           sync.Mutex
	cancelled    bool
}
