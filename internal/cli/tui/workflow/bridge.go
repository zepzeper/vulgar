package workflow

import (
	"encoding/json"
	"fmt"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/engine"
)

// NodeInfo represents a workflow node for TUI display
type NodeInfo struct {
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	Dependencies []string `json:"dependencies"`
	Outputs      []string `json:"outputs"`
	Result       string   `json:"result,omitempty"`
}

// EdgeInfo represents a connection between nodes
type EdgeInfo struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// GraphInfo contains the complete workflow graph structure
type GraphInfo struct {
	Name    string     `json:"name"`
	Status  string     `json:"status"`
	Nodes   []NodeInfo `json:"nodes"`
	Edges   []EdgeInfo `json:"edges"`
	Context string     `json:"context,omitempty"` // Current workflow context (JSON)
}

// Bridge provides a connection between the TUI and workflow engine
type Bridge struct {
	engine     *engine.Engine
	path       string
	workflowUD *lua.LUserData
	graph      GraphInfo
	nodeCode   map[string]NodeCodeInfo // Source code for each node
	editor     EditorConfig
	mu         sync.RWMutex
	loaded     bool
	lastError  string
}

// NewBridge creates a new workflow bridge
func NewBridge() *Bridge {
	return &Bridge{
		editor: DefaultEditor(),
	}
}

// Load loads a workflow file and extracts its graph structure
func (b *Bridge) Load(path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Create a fresh engine for this workflow
	cfg := engine.Config{
		LogLevel:  "WARN",
		LogFormat: "text",
	}
	b.engine = engine.NewEngine(cfg)
	b.path = path
	b.loaded = false
	b.lastError = ""

	// Execute the workflow file to build the graph, but don't run it
	if err := b.engine.L.DoFile(path); err != nil {
		b.lastError = err.Error()
		return fmt.Errorf("failed to load workflow: %w", err)
	}

	// Find the workflow userdata in the global scope
	// Look for common variable names used for workflows
	workflowVarNames := []string{"wf", "workflow", "w"}
	for _, name := range workflowVarNames {
		val := b.engine.L.GetGlobal(name)
		if ud, ok := val.(*lua.LUserData); ok {
			b.workflowUD = ud
			break
		}
	}

	if b.workflowUD == nil {
		// Try to find any workflow userdata in globals
		// This is a fallback if the workflow uses a non-standard variable name
		b.lastError = "no workflow found in script (expected variable 'wf' or 'workflow')"
		return fmt.Errorf("%s", b.lastError)
	}

	// Extract graph info
	if err := b.refreshGraph(); err != nil {
		return err
	}

	// Parse source code for each node
	nodeCode, err := ParseNodeCode(path)
	if err != nil {
		// Non-fatal - just means no code display
		b.nodeCode = make(map[string]NodeCodeInfo)
	} else {
		b.nodeCode = nodeCode
	}

	b.loaded = true
	return nil
}

// refreshGraph extracts the current graph state from the loaded workflow
func (b *Bridge) refreshGraph() error {
	if b.workflowUD == nil {
		return fmt.Errorf("no workflow loaded")
	}

	L := b.engine.L

	// Get workflow status
	L.Push(b.workflowUD)
	statusMethod := L.GetField(L.GetMetatable(b.workflowUD), "__index")
	if statusMethod == lua.LNil {
		return fmt.Errorf("workflow has no methods")
	}

	// Call get_nodes
	L.Push(L.GetField(statusMethod.(*lua.LTable), "get_nodes"))
	L.Push(b.workflowUD)
	if err := L.PCall(1, 1, nil); err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}
	nodesTable := L.Get(-1).(*lua.LTable)
	L.Pop(1)

	// Call get_edges
	L.Push(L.GetField(statusMethod.(*lua.LTable), "get_edges"))
	L.Push(b.workflowUD)
	if err := L.PCall(1, 1, nil); err != nil {
		return fmt.Errorf("failed to get edges: %w", err)
	}
	edgesTable := L.Get(-1).(*lua.LTable)
	L.Pop(1)

	// Call status
	L.Push(L.GetField(statusMethod.(*lua.LTable), "status"))
	L.Push(b.workflowUD)
	if err := L.PCall(1, 1, nil); err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}
	statusTable := L.Get(-1).(*lua.LTable)
	L.Pop(1)

	// Parse nodes
	var nodes []NodeInfo
	nodesTable.ForEach(func(_, v lua.LValue) {
		nodeTable := v.(*lua.LTable)
		node := NodeInfo{
			Name:   lua.LVAsString(nodeTable.RawGetString("name")),
			Status: lua.LVAsString(nodeTable.RawGetString("status")),
		}

		// Parse dependencies
		depsTable := nodeTable.RawGetString("dependencies")
		if deps, ok := depsTable.(*lua.LTable); ok {
			deps.ForEach(func(_, dep lua.LValue) {
				node.Dependencies = append(node.Dependencies, lua.LVAsString(dep))
			})
		}

		// Parse outputs
		outputsTable := nodeTable.RawGetString("outputs")
		if outputs, ok := outputsTable.(*lua.LTable); ok {
			outputs.ForEach(func(_, out lua.LValue) {
				node.Outputs = append(node.Outputs, lua.LVAsString(out))
			})
		}

		// Parse result
		result := nodeTable.RawGetString("result")
		if result != lua.LNil {
			node.Result = luaValueToString(result)
		}

		nodes = append(nodes, node)
	})

	// Parse edges
	var edges []EdgeInfo
	edgesTable.ForEach(func(_, v lua.LValue) {
		edgeTable := v.(*lua.LTable)
		edge := EdgeInfo{
			From: lua.LVAsString(edgeTable.RawGetString("from")),
			To:   lua.LVAsString(edgeTable.RawGetString("to")),
		}
		edges = append(edges, edge)
	})

	// Parse status
	name := lua.LVAsString(statusTable.RawGetString("name"))
	status := lua.LVAsString(statusTable.RawGetString("status"))

	// Get the context (workflow state)
	contextVal := statusTable.RawGetString("context")
	contextStr := ""
	if contextVal != lua.LNil {
		contextStr = luaValueToString(contextVal)
	}

	b.graph = GraphInfo{
		Name:    name,
		Status:  status,
		Nodes:   nodes,
		Edges:   edges,
		Context: contextStr,
	}

	return nil
}

// luaValueToString converts a Lua value to a JSON string
func luaValueToString(v lua.LValue) string {
	switch val := v.(type) {
	case lua.LString:
		return string(val)
	case lua.LNumber:
		return fmt.Sprintf("%v", float64(val))
	case lua.LBool:
		return fmt.Sprintf("%v", bool(val))
	case *lua.LTable:
		result := make(map[string]interface{})
		val.ForEach(func(k, v lua.LValue) {
			key := lua.LVAsString(k)
			result[key] = luaValueToInterface(v)
		})
		b, _ := json.Marshal(result)
		return string(b)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// luaValueToInterface converts a Lua value to a Go interface
func luaValueToInterface(v lua.LValue) interface{} {
	switch val := v.(type) {
	case lua.LString:
		return string(val)
	case lua.LNumber:
		return float64(val)
	case lua.LBool:
		return bool(val)
	case *lua.LTable:
		result := make(map[string]interface{})
		val.ForEach(func(k, v lua.LValue) {
			key := lua.LVAsString(k)
			result[key] = luaValueToInterface(v)
		})
		return result
	default:
		return nil
	}
}

// GetGraph returns the current graph structure
func (b *Bridge) GetGraph() GraphInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.graph
}

// GetNodes returns all node information
func (b *Bridge) GetNodes() []NodeInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.graph.Nodes
}

// GetEdges returns all edge connections
func (b *Bridge) GetEdges() []EdgeInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.graph.Edges
}

// IsLoaded returns whether a workflow is currently loaded
func (b *Bridge) IsLoaded() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.loaded
}

// GetPath returns the loaded workflow path
func (b *Bridge) GetPath() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.path
}

// GetLastError returns the last error message
func (b *Bridge) GetLastError() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.lastError
}

// ExecuteNode runs a single node with dependency resolution
func (b *Bridge) ExecuteNode(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.loaded || b.workflowUD == nil {
		return fmt.Errorf("no workflow loaded")
	}

	L := b.engine.L

	// Get run_node method
	statusMethod := L.GetField(L.GetMetatable(b.workflowUD), "__index")
	if statusMethod == lua.LNil {
		return fmt.Errorf("workflow has no methods")
	}

	// Call run_node(wf, name)
	L.Push(L.GetField(statusMethod.(*lua.LTable), "run_node"))
	L.Push(b.workflowUD)
	L.Push(lua.LString(name))
	if err := L.PCall(2, 2, nil); err != nil {
		return fmt.Errorf("failed to execute node: %w", err)
	}

	// Get results
	errVal := L.Get(-1)
	L.Pop(2)

	if errVal != lua.LNil {
		return fmt.Errorf("%s", lua.LVAsString(errVal))
	}

	// Refresh graph to get updated status
	return b.refreshGraph()
}

// ExecuteAll runs the entire workflow
func (b *Bridge) ExecuteAll() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.loaded || b.workflowUD == nil {
		return fmt.Errorf("no workflow loaded")
	}

	L := b.engine.L

	// Get run method
	statusMethod := L.GetField(L.GetMetatable(b.workflowUD), "__index")
	if statusMethod == lua.LNil {
		return fmt.Errorf("workflow has no methods")
	}

	// Call run(wf, {})
	L.Push(L.GetField(statusMethod.(*lua.LTable), "run"))
	L.Push(b.workflowUD)
	L.Push(L.NewTable()) // Empty input context
	if err := L.PCall(2, 2, nil); err != nil {
		return fmt.Errorf("failed to execute workflow: %w", err)
	}

	// Get results
	errVal := L.Get(-1)
	L.Pop(2)

	if errVal != lua.LNil {
		return fmt.Errorf("%s", lua.LVAsString(errVal))
	}

	// Refresh graph to get updated status
	return b.refreshGraph()
}

// Reset resets all node statuses
func (b *Bridge) Reset() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.loaded || b.workflowUD == nil {
		return fmt.Errorf("no workflow loaded")
	}

	L := b.engine.L

	// Get reset method
	statusMethod := L.GetField(L.GetMetatable(b.workflowUD), "__index")
	if statusMethod == lua.LNil {
		return fmt.Errorf("workflow has no methods")
	}

	// Call reset(wf)
	L.Push(L.GetField(statusMethod.(*lua.LTable), "reset"))
	L.Push(b.workflowUD)
	if err := L.PCall(1, 0, nil); err != nil {
		return fmt.Errorf("failed to reset workflow: %w", err)
	}

	// Refresh graph to get updated status
	return b.refreshGraph()
}

// Close closes the engine and cleans up resources
func (b *Bridge) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.engine != nil {
		b.engine.Close()
		b.engine = nil
	}
	b.workflowUD = nil
	b.loaded = false
}

// GetNodeByName returns a node by its name
func (b *Bridge) GetNodeByName(name string) *NodeInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for i := range b.graph.Nodes {
		if b.graph.Nodes[i].Name == name {
			return &b.graph.Nodes[i]
		}
	}
	return nil
}

// RefreshGraph refreshes the graph state from the workflow (public version)
func (b *Bridge) RefreshGraph() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.refreshGraph()
}

// GetNodeCode returns the source code information for a node
func (b *Bridge) GetNodeCode(name string) *NodeCodeInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if info, exists := b.nodeCode[name]; exists {
		return &info
	}
	return nil
}

// OpenNodeInEditor opens the node's source code in the user's editor
func (b *Bridge) OpenNodeInEditor(name string) error {
	b.mu.RLock()
	codeInfo, exists := b.nodeCode[name]
	path := b.path
	editor := b.editor
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no source code found for node '%s'", name)
	}

	return editor.OpenFile(path, codeInfo.StartLine)
}

// IsTerminalEditor returns true if the configured editor runs in terminal
func (b *Bridge) IsTerminalEditor() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.editor.IsTerminalEditor()
}

// GetEditorName returns the name of the configured editor
func (b *Bridge) GetEditorName() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.editor.Command
}

// Reload reloads the workflow file (for hot-reloading after edits)
func (b *Bridge) Reload() error {
	path := b.GetPath()
	b.Close()
	return b.Load(path)
}
