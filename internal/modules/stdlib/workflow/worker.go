package workflow

import (
	"fmt"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// nodeExecutionResult represents the result of executing a node
type nodeExecutionResult struct {
	nodeName string
	result   lua.LValue
	err      error
}

// nodeExecutionRequest represents a request to execute a node function
type nodeExecutionRequest struct {
	node        *workflowNode
	contextData interface{} // Go data (serialized from Lua table)
	resultChan  chan<- nodeExecutionResult
}

// coordinator runs in a single goroutine and executes all node functions in the main state
// This ensures thread-safety since Lua states aren't thread-safe
func (wf *workflowHandle) coordinator(mainState *lua.LState, execChan <-chan nodeExecutionRequest, done chan<- struct{}) {
	defer close(done)

	for req := range execChan {
		// Check for cancellation
		wf.mu.Lock()
		cancelled := wf.cancelled
		wf.mu.Unlock()

		if cancelled {
			req.resultChan <- nodeExecutionResult{
				nodeName: req.node.name,
				err:      fmt.Errorf("workflow cancelled"),
			}
			continue
		}

		// Convert context data back to Lua table in main state
		mainContext := util.GoToLua(mainState, req.contextData).(*lua.LTable)

		// Execute function in main state (thread-safe - coordinator runs in single goroutine)
		mainState.Push(req.node.fn)
		mainState.Push(mainContext)
		err := mainState.PCall(1, 1, nil)

		var result lua.LValue
		if err != nil {
			req.resultChan <- nodeExecutionResult{
				nodeName: req.node.name,
				err:      fmt.Errorf("node '%s' failed: %w", req.node.name, err),
			}
			continue
		}

		result = mainState.Get(-1)
		mainState.Pop(1)

		// Convert result to Go data for safe transfer
		resultData := util.LuaToGo(result)

		req.resultChan <- nodeExecutionResult{
			nodeName: req.node.name,
			result:   util.GoToLua(mainState, resultData), // Convert back for merging
		}
	}
}

// executeNodesInParallel executes multiple nodes in parallel
// Worker goroutines prepare data, coordinator executes functions in main state
func (wf *workflowHandle) executeNodesInParallel(mainState *lua.LState, nodes []*workflowNode) error {
	if len(nodes) == 0 {
		return nil
	}

	// Channel for execution requests
	execChan := make(chan nodeExecutionRequest, len(nodes))
	// Channel to signal coordinator completion
	coordDone := make(chan struct{})

	// Start coordinator (runs in single goroutine for thread-safety)
	go wf.coordinator(mainState, execChan, coordDone)

	// Channel to collect results
	resultChan := make(chan nodeExecutionResult, len(nodes))

	// Worker goroutines prepare data and send execution requests
	var wg sync.WaitGroup
	for _, node := range nodes {
		wg.Add(1)
		go func(n *workflowNode) {
			defer wg.Done()

			// Mark as running
			n.mu.Lock()
			n.status = NodeStatusRunning
			n.mu.Unlock()

			// Get current context and serialize to Go data
			wf.mu.Lock()
			contextData := util.LuaToGo(wf.context)
			wf.mu.Unlock()

			// Send execution request to coordinator
			execChan <- nodeExecutionRequest{
				node:        n,
				contextData: contextData,
				resultChan:  resultChan,
			}
		}(node)
	}

	// Close execChan when all workers are done
	go func() {
		wg.Wait()
		close(execChan)
	}()

	// Collect all results
	var firstError error
	resultsReceived := 0
	totalNodes := len(nodes)

	// Collect results until we have all of them
	for resultsReceived < totalNodes {
		result := <-resultChan
		resultsReceived++

		// Update node status
		node := wf.nodes[result.nodeName]
		if node != nil {
			node.mu.Lock()
			if result.err != nil {
				node.status = NodeStatusFailed
			} else {
				node.result = result.result
				node.status = NodeStatusCompleted
			}
			node.mu.Unlock()
		}

		if result.err != nil && firstError == nil {
			firstError = result.err
			// Continue collecting other results even if one failed
			continue
		}

		if result.err == nil {
			// Merge result into shared context
			wf.mergeContext(result.result)
		}
	}

	// Wait for coordinator to finish
	<-coordDone

	return firstError
}