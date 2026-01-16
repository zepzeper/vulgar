package workflow

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

func (wf *workflowHandle) getReadyNodes() []*workflowNode {
	wf.mu.Lock()
	defer wf.mu.Unlock()

	var ready []*workflowNode
	for _, node := range wf.nodes {
		// Skip completed, failed or running
		if node.status != NodeStatusPending {
			continue
		}

		// Check if all dependencies are completed
		allDepsReady := true
		for _, depName := range node.dependencies {
			depNode, exists := wf.nodes[depName]
			if !exists || depNode.status != NodeStatusCompleted {
				allDepsReady = false
				break
			}
		}

		if allDepsReady {
			ready = append(ready, node)
		}
	}

	return ready
}

func (wf *workflowHandle) mergeContext(result lua.LValue) {
	if result == nil {
		return
	}

	wf.mu.Lock()
	defer wf.mu.Unlock()

	// If result is a table, merge its contents into context
	if tbl, ok := result.(*lua.LTable); ok {
		tbl.ForEach(func(k, v lua.LValue) {
			wf.context.RawSet(k, v)
		})
	}
}

func (wf *workflowHandle) executeNode(L *lua.LState, node *workflowNode) error {
	// Mark as running
	node.mu.Lock()
	node.status = NodeStatusRunning
	node.mu.Unlock()

	// Call the node's function with the current context
	L.Push(node.fn)
	L.Push(wf.context)
	if err := L.PCall(1, 1, nil); err != nil {
		node.mu.Lock()
		node.status = NodeStatusFailed
		node.mu.Unlock()
		return fmt.Errorf("node '%s' failed: %w", node.name, err)
	}

	// Get result
	result := L.Get(-1)
	L.Pop(1)

	// Store result and mark as completed
	node.mu.Lock()
	node.result = result
	node.status = NodeStatusCompleted
	node.mu.Unlock()

	// Merge result into shared context
	wf.mergeContext(result)

	return nil
}

func (wf *workflowHandle) executeGraph(L *lua.LState) error {
	// Reset all node statuses
	wf.mu.Lock()
	for _, node := range wf.nodes {
		node.status = NodeStatusPending
		node.result = nil
	}
	wf.mu.Unlock()

	completedCount := 0
	totalNodes := len(wf.nodes)

	// Main execution loop
	for completedCount < totalNodes {
		// Check for cancellation
		wf.mu.Lock()
		if wf.cancelled {
			wf.status = WorkflowStatusCancelled
			wf.mu.Unlock()
			return fmt.Errorf("workflow cancelled")
		}
		wf.mu.Unlock()

		// Get all ready nodes (nodes with satisfied dependencies)
		readyNodes := wf.getReadyNodes()

		if len(readyNodes) == 0 {
			// No ready nodes - check if we're stuck (circular dependency or error)
			wf.mu.Lock()
			hasPending := false
			for _, node := range wf.nodes {
				if node.status == NodeStatusPending {
					hasPending = true
					break
				}
			}
			wf.mu.Unlock()

			if hasPending {
				return fmt.Errorf("deadlock detected: nodes have unsatisfied dependencies")
			}
			break
		}

		// Execute ready nodes in parallel
		// Worker goroutines prepare data, coordinator executes functions in main state
		if err := wf.executeNodesInParallel(L, readyNodes); err != nil {
			wf.mu.Lock()
			wf.status = WorkflowStatusFailed
			wf.mu.Unlock()
			return err
		}

		// Count completed nodes
		wf.mu.Lock()
		for _, node := range readyNodes {
			if node.status == NodeStatusCompleted {
				completedCount++
			}
		}
		wf.mu.Unlock()
	}

	// All nodes completed successfully
	wf.mu.Lock()
	wf.status = WorkflowStatusCompleted
	// Set final output to the last completed node's result (or merge all?)
	wf.mu.Unlock()

	return nil
}

// luaWorkflowRun executes the workflow graph
func luaWorkflowRun(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return util.PushError(L, "workflow is required")
	}

	wf := checkWorkflow(L, 1)
	if wf == nil {
		return util.PushError(L, "invalid workflow")
	}

	input := L.OptTable(2, L.NewTable())

	wf.mu.Lock()
	if wf.status == WorkflowStatusRunning {
		wf.mu.Unlock()
		return util.PushError(L, "workflow is already running")
	}

	if len(wf.nodes) == 0 {
		wf.mu.Unlock()
		return util.PushError(L, "workflow has no nodes")
	}

	wf.status = WorkflowStatusRunning
	wf.cancelled = false
	wf.context = input
	errorHandler := wf.errorHandler
	wf.mu.Unlock()

	// Execute the graph
	err := wf.executeGraph(L)

	if err != nil {
		wf.mu.Lock()
		wf.status = WorkflowStatusFailed
		wf.mu.Unlock()

		// Call error handler if set
		if errorHandler != nil {
			L.Push(errorHandler)
			L.Push(lua.LString(err.Error()))
			_ = L.PCall(1, 0, nil) // Ignore error handler errors
		}

		return util.PushError(L, err.Error())
	}

	// Return final context as result
	wf.mu.Lock()
	result := wf.context
	wf.mu.Unlock()

	return util.PushSuccess(L, result)
}
