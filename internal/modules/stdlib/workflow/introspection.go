package workflow

import (
	lua "github.com/yuin/gopher-lua"
)

// luaGetNodes returns a table with all node information for TUI inspection
// Usage: local nodes = workflow.get_nodes(wf)
// Returns: { {name="node1", status="pending", dependencies={"dep1"}, result=...}, ... }
func luaGetNodes(L *lua.LState) int {
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
	defer wf.mu.Unlock()

	result := L.NewTable()
	idx := 1

	for name, node := range wf.nodes {
		nodeTable := L.NewTable()
		nodeTable.RawSetString("name", lua.LString(name))
		nodeTable.RawSetString("status", lua.LString(node.status))

		// Add dependencies
		depsTable := L.NewTable()
		for i, dep := range node.dependencies {
			depsTable.RawSetInt(i+1, lua.LString(dep))
		}
		nodeTable.RawSetString("dependencies", depsTable)

		// Add outputs (nodes that depend on this one)
		outputsTable := L.NewTable()
		for i, out := range node.outputs {
			outputsTable.RawSetInt(i+1, lua.LString(out))
		}
		nodeTable.RawSetString("outputs", outputsTable)

		// Add result if available
		if node.result != nil {
			nodeTable.RawSetString("result", node.result)
		} else {
			nodeTable.RawSetString("result", lua.LNil)
		}

		result.RawSetInt(idx, nodeTable)
		idx++
	}

	L.Push(result)
	return 1
}

// luaGetEdges returns a table with all edges (connections) for TUI visualization
// Usage: local edges = workflow.get_edges(wf)
// Returns: { {from="node1", to="node2"}, ... }
func luaGetEdges(L *lua.LState) int {
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
	defer wf.mu.Unlock()

	result := L.NewTable()
	idx := 1

	// Build edges from dependencies
	for nodeName, node := range wf.nodes {
		for _, depName := range node.dependencies {
			edgeTable := L.NewTable()
			edgeTable.RawSetString("from", lua.LString(depName))
			edgeTable.RawSetString("to", lua.LString(nodeName))
			result.RawSetInt(idx, edgeTable)
			idx++
		}
	}

	L.Push(result)
	return 1
}

// luaRunNode executes a single node with dependency resolution
// Usage: local result, err = workflow.run_node(wf, "node_name")
// If dependencies haven't completed, runs them first
func luaRunNode(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LNil)
		L.Push(lua.LString("workflow is required"))
		return 2
	}

	wf := checkWorkflow(L, 1)
	if wf == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("invalid workflow"))
		return 2
	}

	nodeName := L.CheckString(2)

	wf.mu.Lock()
	node, exists := wf.nodes[nodeName]
	if !exists {
		wf.mu.Unlock()
		L.Push(lua.LNil)
		L.Push(lua.LString("node '" + nodeName + "' does not exist"))
		return 2
	}

	// Initialize context if nil
	if wf.context == nil {
		wf.context = L.NewTable()
	}
	wf.mu.Unlock()

	// Execute dependencies first (recursive)
	if err := wf.executeDependencies(L, node); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Execute the node itself
	if err := wf.executeNode(L, node); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Return node result
	node.mu.Lock()
	result := node.result
	node.mu.Unlock()

	if result != nil {
		L.Push(result)
	} else {
		L.Push(lua.LNil)
	}
	L.Push(lua.LNil)
	return 2
}

// executeDependencies recursively executes all dependencies of a node
func (wf *workflowHandle) executeDependencies(L *lua.LState, node *workflowNode) error {
	for _, depName := range node.dependencies {
		wf.mu.Lock()
		depNode, exists := wf.nodes[depName]
		wf.mu.Unlock()

		if !exists {
			continue
		}

		depNode.mu.Lock()
		status := depNode.status
		depNode.mu.Unlock()

		// Skip if already completed
		if status == NodeStatusCompleted {
			continue
		}

		// Execute dependency's dependencies first
		if err := wf.executeDependencies(L, depNode); err != nil {
			return err
		}

		// Execute the dependency node
		if err := wf.executeNode(L, depNode); err != nil {
			return err
		}
	}
	return nil
}

// luaReset resets all node statuses to pending
// Usage: workflow.reset(wf)
func luaReset(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return 0
	}

	wf := checkWorkflow(L, 1)
	if wf == nil {
		return 0
	}

	wf.mu.Lock()
	defer wf.mu.Unlock()

	for _, node := range wf.nodes {
		node.mu.Lock()
		node.status = NodeStatusPending
		node.result = nil
		node.mu.Unlock()
	}

	wf.status = WorkflowStatusPending
	wf.cancelled = false
	wf.context = L.NewTable()

	return 0
}

// luaGetNodeStatus returns the status of a specific node
// Usage: local status = workflow.get_node_status(wf, "node_name")
func luaGetNodeStatus(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LNil)
		return 1
	}

	wf := checkWorkflow(L, 1)
	if wf == nil {
		L.Push(lua.LNil)
		return 1
	}

	nodeName := L.CheckString(2)

	wf.mu.Lock()
	node, exists := wf.nodes[nodeName]
	wf.mu.Unlock()

	if !exists {
		L.Push(lua.LNil)
		return 1
	}

	node.mu.Lock()
	status := node.status
	result := node.result
	node.mu.Unlock()

	statusTable := L.NewTable()
	statusTable.RawSetString("name", lua.LString(nodeName))
	statusTable.RawSetString("status", lua.LString(status))
	if result != nil {
		statusTable.RawSetString("result", result)
	}

	L.Push(statusTable)
	return 1
}
