package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zepzeper/vulgar/internal/cli"
	"github.com/zepzeper/vulgar/internal/cli/tui/workflow"
)

// WorkflowInspector is the TUI component for inspecting and executing workflows
type WorkflowInspector struct {
	width         int
	height        int
	bridge        *workflow.Bridge
	renderer      *GraphRenderer
	path          string
	selectedIndex int
	nodeNames     []string
	loading       bool
	error         string
	executing     bool
	lastResult    string
	showHelp      bool
	showCode      bool
}

// NewWorkflowInspector creates a new workflow inspector
func NewWorkflowInspector(path string) *WorkflowInspector {
	return &WorkflowInspector{
		path:     path,
		bridge:   workflow.NewBridge(),
		renderer: NewGraphRenderer(),
		loading:  true,
	}
}

// workflowLoadedMsg is sent when workflow loading completes
type workflowLoadedMsg struct {
	err error
}

// nodeExecutedMsg is sent when a node execution completes
type nodeExecutedMsg struct {
	nodeName string
	err      error
}

// workflowExecutedMsg is sent when full workflow execution completes
type workflowExecutedMsg struct {
	err error
}

// workflowResetMsg is sent when workflow reset completes
type workflowResetMsg struct {
	err error
}

// editorOpenedMsg is sent when editor open completes
type editorOpenedMsg struct {
	nodeName string
	err      error
}

// workflowReloadedMsg is sent when workflow reload completes
type workflowReloadedMsg struct {
	err error
}

// Init initializes the inspector
func (i *WorkflowInspector) Init() tea.Cmd {
	return i.loadWorkflow
}

// loadWorkflow is a tea.Cmd that loads the workflow
func (i *WorkflowInspector) loadWorkflow() tea.Msg {
	err := i.bridge.Load(i.path)
	return workflowLoadedMsg{err: err}
}

// executeNode is a tea.Cmd that executes a single node
func (i *WorkflowInspector) executeNode(name string) tea.Cmd {
	return func() tea.Msg {
		err := i.bridge.ExecuteNode(name)
		return nodeExecutedMsg{nodeName: name, err: err}
	}
}

// executeAll is a tea.Cmd that executes the entire workflow
func (i *WorkflowInspector) executeAll() tea.Cmd {
	return func() tea.Msg {
		err := i.bridge.ExecuteAll()
		return workflowExecutedMsg{err: err}
	}
}

// resetWorkflow is a tea.Cmd that resets the workflow
func (i *WorkflowInspector) resetWorkflow() tea.Cmd {
	return func() tea.Msg {
		err := i.bridge.Reset()
		return workflowResetMsg{err: err}
	}
}

// openInEditor is a tea.Cmd that opens the node in the user's editor
func (i *WorkflowInspector) openInEditor(nodeName string) tea.Cmd {
	return func() tea.Msg {
		err := i.bridge.OpenNodeInEditor(nodeName)
		return editorOpenedMsg{nodeName: nodeName, err: err}
	}
}

// reloadWorkflow is a tea.Cmd that reloads the workflow file
func (i *WorkflowInspector) reloadWorkflow() tea.Cmd {
	return func() tea.Msg {
		err := i.bridge.Reload()
		return workflowReloadedMsg{err: err}
	}
}

// Update handles messages
func (i *WorkflowInspector) Update(msg tea.Msg) (*WorkflowInspector, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case workflowLoadedMsg:
		i.loading = false
		if msg.err != nil {
			i.error = msg.err.Error()
		} else {
			i.error = ""
			i.refreshNodeList()
		}

	case nodeExecutedMsg:
		i.executing = false
		if msg.err != nil {
			i.lastResult = fmt.Sprintf("Error executing %s: %s", msg.nodeName, msg.err.Error())
		} else {
			i.lastResult = fmt.Sprintf("Node '%s' executed successfully", msg.nodeName)
			i.refreshNodeList()
		}

	case workflowExecutedMsg:
		i.executing = false
		if msg.err != nil {
			i.lastResult = fmt.Sprintf("Error: %s", msg.err.Error())
		} else {
			i.lastResult = "Workflow completed successfully"
			i.refreshNodeList()
		}

	case workflowResetMsg:
		i.executing = false
		if msg.err != nil {
			i.lastResult = fmt.Sprintf("Error resetting: %s", msg.err.Error())
		} else {
			i.lastResult = "Workflow reset"
			i.refreshNodeList()
		}

	case editorOpenedMsg:
		if msg.err != nil {
			i.lastResult = fmt.Sprintf("Error opening editor: %s", msg.err.Error())
		} else {
			i.lastResult = fmt.Sprintf("Opened %s in editor. Press L to reload after editing.", msg.nodeName)
		}

	case workflowReloadedMsg:
		i.loading = false
		if msg.err != nil {
			i.lastResult = fmt.Sprintf("Error reloading: %s", msg.err.Error())
		} else {
			i.lastResult = "Workflow reloaded successfully"
			i.refreshNodeList()
		}

	case tea.WindowSizeMsg:
		i.width = msg.Width
		i.height = msg.Height
		i.renderer.SetDimensions(msg.Width, msg.Height)

	case tea.KeyMsg:
		if i.loading || i.executing {
			return i, nil
		}

		switch msg.String() {
		case "up", "k":
			i.navigateUp()
		case "down", "j":
			i.navigateDown()
		case "left", "h":
			i.navigatePrev()
		case "right", "l":
			i.navigateNext()
		case "enter":
			if i.selectedIndex >= 0 && i.selectedIndex < len(i.nodeNames) {
				i.executing = true
				i.lastResult = fmt.Sprintf("Executing %s...", i.nodeNames[i.selectedIndex])
				return i, i.executeNode(i.nodeNames[i.selectedIndex])
			}
		case "r", "R":
			i.executing = true
			i.lastResult = "Running entire workflow..."
			return i, i.executeAll()
		case "x", "X":
			i.executing = true
			i.lastResult = "Resetting workflow..."
			return i, i.resetWorkflow()
		case "e", "E":
			// Open current node in editor
			if i.selectedIndex >= 0 && i.selectedIndex < len(i.nodeNames) {
				nodeName := i.nodeNames[i.selectedIndex]
				editorName := i.bridge.GetEditorName()

				// For terminal editors, we can't run them while TUI is active
				// Show message about which editor will be used
				if i.bridge.IsTerminalEditor() {
					i.lastResult = fmt.Sprintf("Opening %s in %s... (TUI will pause)", nodeName, editorName)
				} else {
					i.lastResult = fmt.Sprintf("Opening %s in %s...", nodeName, editorName)
				}

				return i, i.openInEditor(nodeName)
			}
		case "L":
			// Reload workflow (hot reload)
			i.loading = true
			i.lastResult = "Reloading workflow..."
			return i, i.reloadWorkflow()
		case "c", "C":
			// Toggle code view
			i.showCode = !i.showCode
		case "?":
			i.showHelp = !i.showHelp
		}
	}

	return i, cmd
}

// refreshNodeList updates the node name list from the bridge
func (i *WorkflowInspector) refreshNodeList() {
	nodes := i.bridge.GetNodes()
	i.nodeNames = make([]string, len(nodes))
	for idx, node := range nodes {
		i.nodeNames[idx] = node.Name
	}
	// Update selected node in renderer
	if i.selectedIndex >= 0 && i.selectedIndex < len(i.nodeNames) {
		i.renderer.SetSelected(i.nodeNames[i.selectedIndex])
	} else if len(i.nodeNames) > 0 {
		i.selectedIndex = 0
		i.renderer.SetSelected(i.nodeNames[0])
	}
}

// navigateUp moves selection up (to previous layer in graph)
func (i *WorkflowInspector) navigateUp() {
	if len(i.nodeNames) == 0 {
		return
	}
	// Find node in previous layer (with fewer dependencies)
	currentNode := i.bridge.GetNodeByName(i.nodeNames[i.selectedIndex])
	if currentNode != nil && len(currentNode.Dependencies) > 0 {
		// Move to first dependency
		for idx, name := range i.nodeNames {
			if name == currentNode.Dependencies[0] {
				i.selectedIndex = idx
				i.renderer.SetSelected(name)
				return
			}
		}
	}
}

// navigateDown moves selection down (to next layer in graph)
func (i *WorkflowInspector) navigateDown() {
	if len(i.nodeNames) == 0 {
		return
	}
	// Find node in next layer (that depends on current)
	currentNode := i.bridge.GetNodeByName(i.nodeNames[i.selectedIndex])
	if currentNode != nil && len(currentNode.Outputs) > 0 {
		// Move to first output
		for idx, name := range i.nodeNames {
			if name == currentNode.Outputs[0] {
				i.selectedIndex = idx
				i.renderer.SetSelected(name)
				return
			}
		}
	}
}

// navigatePrev moves to previous node in list
func (i *WorkflowInspector) navigatePrev() {
	if len(i.nodeNames) == 0 {
		return
	}
	i.selectedIndex--
	if i.selectedIndex < 0 {
		i.selectedIndex = len(i.nodeNames) - 1
	}
	i.renderer.SetSelected(i.nodeNames[i.selectedIndex])
}

// navigateNext moves to next node in list
func (i *WorkflowInspector) navigateNext() {
	if len(i.nodeNames) == 0 {
		return
	}
	i.selectedIndex++
	if i.selectedIndex >= len(i.nodeNames) {
		i.selectedIndex = 0
	}
	i.renderer.SetSelected(i.nodeNames[i.selectedIndex])
}

// View renders the inspector
func (i *WorkflowInspector) View(width, height int) string {
	i.width = width
	i.height = height
	i.renderer.SetDimensions(width, height)

	contentWidth := width - 4
	contentHeight := height - 6

	if contentWidth < 0 {
		contentWidth = 0
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	var content string

	if i.loading {
		content = i.renderLoading()
	} else if i.error != "" {
		content = i.renderError()
	} else if i.showHelp {
		content = i.renderHelpScreen()
	} else {
		content = i.renderInspector()
	}

	return lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Padding(1, 2).
		Render(content)
}

// renderLoading renders the loading state
func (i *WorkflowInspector) renderLoading() string {
	return cli.Title("Loading Workflow") + "\n\n" +
		cli.Muted("Path: "+i.path) + "\n\n" +
		cli.Info("Loading...")
}

// renderError renders the error state
func (i *WorkflowInspector) renderError() string {
	return cli.Title("Workflow Error") + "\n\n" +
		cli.Muted("Path: "+i.path) + "\n\n" +
		cli.Error("Error: "+i.error) + "\n\n" +
		cli.Muted("Press Esc to go back")
}

// renderHelpScreen renders the full help screen
func (i *WorkflowInspector) renderHelpScreen() string {
	help := []string{
		cli.Title("Workflow Inspector Help"),
		"",
		cli.Info("Navigation:"),
		"  ↑/k       Navigate to parent node (dependency)",
		"  ↓/j       Navigate to child node (downstream)",
		"  ←/h       Previous node in list",
		"  →/l       Next node in list",
		"",
		cli.Info("Execution:"),
		"  Enter     Execute selected node (with dependencies)",
		"  R         Run entire workflow",
		"  X         Reset all node statuses",
		"",
		cli.Info("Editing:"),
		"  e         Open node in editor ($EDITOR)",
		"  L         Reload workflow (hot reload)",
		"",
		cli.Info("View:"),
		"  c         Toggle code view",
		"  ?         Toggle this help screen",
		"  Esc       Return to workflow list",
		"",
		cli.Muted("Editor: " + i.bridge.GetEditorName()),
		cli.Muted("Press ? to close help"),
	}
	return strings.Join(help, "\n")
}

// renderInspector renders the main inspector view
func (i *WorkflowInspector) renderInspector() string {
	graph := i.bridge.GetGraph()

	// Calculate layout
	graphHeight := i.height - 15 // Reserve space for detail and status
	if graphHeight < 5 {
		graphHeight = 5
	}

	var sections []string

	// Header
	pathStyle := lipgloss.NewStyle().Foreground(cli.ColorMuted)
	sections = append(sections, pathStyle.Render("File: "+i.path))
	sections = append(sections, "")

	// Graph view
	graphView := i.renderer.Render(graph)
	sections = append(sections, graphView)
	sections = append(sections, "")

	// Separator
	separatorStyle := lipgloss.NewStyle().
		Foreground(cli.ColorMuted).
		Width(i.width - 8)
	sections = append(sections, separatorStyle.Render(strings.Repeat("─", i.width-10)))

	// Node detail panel
	var selectedNode *workflow.NodeInfo
	var selectedCode *workflow.NodeCodeInfo
	if i.selectedIndex >= 0 && i.selectedIndex < len(i.nodeNames) {
		nodeName := i.nodeNames[i.selectedIndex]
		selectedNode = i.bridge.GetNodeByName(nodeName)
		selectedCode = i.bridge.GetNodeCode(nodeName)
	}
	detailView := i.renderer.RenderNodeDetail(selectedNode, graph.Context)
	sections = append(sections, detailView)

	// Code view (if enabled)
	if i.showCode && selectedCode != nil {
		sections = append(sections, "")
		codeView := i.renderer.RenderNodeCode(selectedCode)
		sections = append(sections, codeView)
	}

	// Status/result line
	if i.lastResult != "" {
		sections = append(sections, "")
		if strings.HasPrefix(i.lastResult, "Error") {
			sections = append(sections, cli.Error(i.lastResult))
		} else if strings.HasPrefix(i.lastResult, "Executing") || strings.HasPrefix(i.lastResult, "Running") || strings.HasPrefix(i.lastResult, "Resetting") || strings.HasPrefix(i.lastResult, "Reloading") {
			sections = append(sections, cli.Info(i.lastResult))
		} else {
			sections = append(sections, cli.Success(i.lastResult))
		}
	}

	// Help line
	sections = append(sections, "")
	sections = append(sections, i.renderer.RenderHelp())

	return strings.Join(sections, "\n")
}

// Close cleans up resources
func (i *WorkflowInspector) Close() {
	if i.bridge != nil {
		i.bridge.Close()
	}
}

// ShouldExit returns true if the inspector should exit (Esc pressed)
func (i *WorkflowInspector) ShouldExit() bool {
	return false // This is handled by the parent workflows tab
}
