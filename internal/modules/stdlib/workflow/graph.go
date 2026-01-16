package workflow

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

func (wf *workflowHandle) computeOutputs() {
	wf.mu.Lock()
	defer wf.mu.Unlock()

	for _, node := range wf.nodes {
		node.outputs = []string{}
	}

	for nodeName, node := range wf.nodes {
		for _, depName := range node.dependencies {
			if depNode, exists := wf.nodes[depName]; exists {
				found := false

				for _, out := range depNode.outputs {
					if out == nodeName {
						found = true
						break
					}
				}

				if !found {
					depNode.outputs = append(depNode.outputs, nodeName)
				}
			}
		}
	}

}

// Usage: local err = workflow.node(wf, "node_name", function(ctx) return result end)
// Usage: local err = workflow.node(wf, "node_name", function(ctx) return result end, {depends_on = {"node1", "node2"}})
func luaNode(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return util.PushError(L, "workflow is required")
	}

	wf := checkWorkflow(L, 1)
	if wf == nil {
		return util.PushError(L, "invalid workflow")
	}

	name := L.CheckString(2)
	fn := L.CheckFunction(3)
	opts := L.OptTable(4, nil)

	// Extract dependencies from options
	var dependencies []string
	if opts != nil {
		if depsTable := L.GetField(opts, "depends_on"); depsTable != lua.LNil {
			if deps, ok := depsTable.(*lua.LTable); ok {
				deps.ForEach(func(_, v lua.LValue) {
					if depName, ok := v.(lua.LString); ok {
						dependencies = append(dependencies, string(depName))
					}
				})
			}
		}
	}

	wf.mu.Lock()
	defer wf.mu.Unlock()

	if _, exists := wf.nodes[name]; exists {
		return util.PushError(L, "node '%s' already exists", name)
	}

	// Validate dependencies exist
	for _, depName := range dependencies {
		if _, exists := wf.nodes[depName]; !exists {
			return util.PushError(L, "dependency '%s' does not exist for node '%s'", depName, name)
		}
	}

	// Create and add node
	node := &workflowNode{
		name:         name,
		fn:           fn,
		dependencies: dependencies,
		outputs:      []string{},
		status:       NodeStatusPending,
	}

	if wf.nodes == nil {
		wf.nodes = make(map[string]*workflowNode)
	}
	wf.nodes[name] = node

	// Recompute outputs for all nodes
	wf.computeOutputs()

	L.Push(lua.LNil)
	return 1
}

// Usage: local err = workflow.edge(wf, "from_node", "to_node")
// Alternative way to connect nodes (adds dependency)
func luaEdge(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return util.PushError(L, "workflow is required")
	}

	wf := checkWorkflow(L, 1)
	if wf == nil {
		return util.PushError(L, "invalid workflow")
	}

	fromName := L.CheckString(2)
	toName := L.CheckString(3)

	wf.mu.Lock()
	defer wf.mu.Unlock()

	// Validate nodes exist
	toNode, exists := wf.nodes[toName]
	if !exists {
		return util.PushError(L, "node '%s' does not exist", toName)
	}
	if _, exists := wf.nodes[fromName]; !exists {
		return util.PushError(L, "node '%s' does not exist", fromName)
	}

	// Check if dependency already exists
	for _, dep := range toNode.dependencies {
		if dep == fromName {
			// Edge already exists, no error
			L.Push(lua.LNil)
			return 1
		}
	}

	// Add dependency
	toNode.dependencies = append(toNode.dependencies, fromName)

	// Recompute outputs
	wf.computeOutputs()

	L.Push(lua.LNil)
	return 1
}
