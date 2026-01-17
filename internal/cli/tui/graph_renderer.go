package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zepzeper/vulgar/internal/cli"
	"github.com/zepzeper/vulgar/internal/cli/tui/workflow"
)

// GraphRenderer renders workflow graphs as ASCII art
type GraphRenderer struct {
	width      int
	height     int
	nodeWidth  int
	nodeHeight int
	selected   string
}

// NewGraphRenderer creates a new graph renderer
func NewGraphRenderer() *GraphRenderer {
	return &GraphRenderer{
		nodeWidth:  14,
		nodeHeight: 3,
	}
}

// SetDimensions sets the available rendering area
func (r *GraphRenderer) SetDimensions(width, height int) {
	r.width = width
	r.height = height
}

// SetSelected sets the currently selected node
func (r *GraphRenderer) SetSelected(name string) {
	r.selected = name
}

// statusIcon returns an icon for the given status
func statusIcon(status string) string {
	switch status {
	case "pending":
		return "○"
	case "running":
		return "●"
	case "completed":
		return "✓"
	case "failed":
		return "✗"
	case "skipped":
		return "⊘"
	default:
		return "?"
	}
}

// statusColor returns a color for the given status
func statusColor(status string) lipgloss.Color {
	switch status {
	case "pending":
		return cli.ColorMuted
	case "running":
		return lipgloss.Color("#FFD700") // Gold/Yellow
	case "completed":
		return cli.ColorSecondary
	case "failed":
		return cli.ColorError
	case "skipped":
		return cli.ColorMuted
	default:
		return cli.ColorMuted
	}
}

// Render renders the workflow graph as ASCII art with horizontal layout
func (r *GraphRenderer) Render(graph workflow.GraphInfo) string {
	if len(graph.Nodes) == 0 {
		return cli.Muted("No nodes in workflow")
	}

	// Build graph layout
	layers := r.buildLayers(graph)

	var lines []string

	// Render workflow header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(cli.ColorPrimary)
	statusStyle := lipgloss.NewStyle().
		Foreground(statusColor(graph.Status))

	header := fmt.Sprintf("%s  %s %s",
		headerStyle.Render(graph.Name),
		statusIcon(graph.Status),
		statusStyle.Render(graph.Status))
	lines = append(lines, header, "")

	// Render horizontal flow: layer1 → layer2 → layer3 → ...
	flowLine := r.renderHorizontalFlow(layers)
	lines = append(lines, flowLine)

	return strings.Join(lines, "\n")
}

// buildLayers organizes nodes into layers based on dependencies (topological sort)
func (r *GraphRenderer) buildLayers(graph workflow.GraphInfo) [][]workflow.NodeInfo {
	// Build dependency map
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)
	nodeMap := make(map[string]workflow.NodeInfo)

	for _, node := range graph.Nodes {
		nodeMap[node.Name] = node
		if _, exists := inDegree[node.Name]; !exists {
			inDegree[node.Name] = 0
		}
		for _, dep := range node.Dependencies {
			dependents[dep] = append(dependents[dep], node.Name)
			inDegree[node.Name]++
		}
	}

	// Find root nodes (no dependencies)
	var layers [][]workflow.NodeInfo
	var currentLayer []workflow.NodeInfo

	for _, node := range graph.Nodes {
		if inDegree[node.Name] == 0 {
			currentLayer = append(currentLayer, node)
		}
	}

	// Sort current layer for consistent display
	sort.Slice(currentLayer, func(i, j int) bool {
		return currentLayer[i].Name < currentLayer[j].Name
	})

	// Process layers
	processed := make(map[string]bool)
	for len(currentLayer) > 0 {
		layers = append(layers, currentLayer)

		for _, node := range currentLayer {
			processed[node.Name] = true
		}

		var nextLayer []workflow.NodeInfo
		for _, node := range currentLayer {
			for _, depName := range dependents[node.Name] {
				if processed[depName] {
					continue
				}
				// Check if all dependencies are processed
				depNode := nodeMap[depName]
				allDepsProcessed := true
				for _, dep := range depNode.Dependencies {
					if !processed[dep] {
						allDepsProcessed = false
						break
					}
				}
				if allDepsProcessed {
					// Check if already in next layer
					found := false
					for _, n := range nextLayer {
						if n.Name == depName {
							found = true
							break
						}
					}
					if !found {
						nextLayer = append(nextLayer, depNode)
					}
				}
			}
		}

		// Sort next layer for consistent display
		sort.Slice(nextLayer, func(i, j int) bool {
			return nextLayer[i].Name < nextLayer[j].Name
		})

		currentLayer = nextLayer
	}

	return layers
}

// renderHorizontalFlow renders all nodes in a horizontal flow with arrows
func (r *GraphRenderer) renderHorizontalFlow(layers [][]workflow.NodeInfo) string {
	if len(layers) == 0 {
		return ""
	}

	// Node box styles
	normalBorder := lipgloss.RoundedBorder()
	selectedBorder := lipgloss.ThickBorder()

	// Render each node and collect rendered boxes
	var renderedNodes []string
	var allNodes []workflow.NodeInfo

	for _, layer := range layers {
		for _, node := range layer {
			allNodes = append(allNodes, node)
		}
	}

	for _, node := range allNodes {
		isSelected := node.Name == r.selected

		var boxStyle lipgloss.Style
		if isSelected {
			boxStyle = lipgloss.NewStyle().
				Border(selectedBorder).
				BorderForeground(cli.ColorPrimary).
				Width(r.nodeWidth - 2).
				Align(lipgloss.Center)
		} else {
			boxStyle = lipgloss.NewStyle().
				Border(normalBorder).
				BorderForeground(statusColor(node.Status)).
				Width(r.nodeWidth - 2).
				Align(lipgloss.Center)
		}

		// Truncate name if needed
		name := node.Name
		maxNameLen := r.nodeWidth - 4
		if len(name) > maxNameLen {
			name = name[:maxNameLen-2] + ".."
		}

		// Status line with icon
		statusLine := fmt.Sprintf("%s %s", statusIcon(node.Status), node.Status)
		if len(statusLine) > maxNameLen {
			statusLine = statusLine[:maxNameLen]
		}

		// Create node content
		content := fmt.Sprintf("%s\n%s",
			lipgloss.NewStyle().Bold(true).Render(name),
			lipgloss.NewStyle().Foreground(statusColor(node.Status)).Render(statusLine))

		box := boxStyle.Render(content)
		renderedNodes = append(renderedNodes, box)
	}

	// Join nodes with arrows horizontally
	arrow := lipgloss.NewStyle().Foreground(cli.ColorMuted).Render(" ─▶ ")

	// We need to join multi-line boxes side by side
	if len(renderedNodes) == 0 {
		return ""
	}

	// Split each node into lines
	nodeLines := make([][]string, len(renderedNodes))
	maxLines := 0
	for i, node := range renderedNodes {
		nodeLines[i] = strings.Split(node, "\n")
		if len(nodeLines[i]) > maxLines {
			maxLines = len(nodeLines[i])
		}
	}

	// Find the middle line for the arrow
	arrowLine := maxLines / 2

	// Build output line by line
	var resultLines []string
	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		var lineBuilder strings.Builder
		for nodeIdx, lines := range nodeLines {
			// Add the line from this node (or spaces if the node has fewer lines)
			if lineIdx < len(lines) {
				lineBuilder.WriteString(lines[lineIdx])
			} else {
				// Pad with spaces to match node width
				lineBuilder.WriteString(strings.Repeat(" ", r.nodeWidth))
			}

			// Add arrow between nodes (only on middle line)
			if nodeIdx < len(nodeLines)-1 {
				if lineIdx == arrowLine {
					lineBuilder.WriteString(arrow)
				} else {
					lineBuilder.WriteString("     ") // Same width as arrow
				}
			}
		}
		resultLines = append(resultLines, lineBuilder.String())
	}

	return strings.Join(resultLines, "\n")
}

// RenderNodeDetail renders detailed information about a selected node
func (r *GraphRenderer) RenderNodeDetail(node *workflow.NodeInfo, inputContext string) string {
	if node == nil {
		return cli.Muted("No node selected")
	}

	var lines []string

	// Node name with status icon
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(cli.ColorPrimary)
	statusStyle := lipgloss.NewStyle().Foreground(statusColor(node.Status))
	lines = append(lines, fmt.Sprintf("%s  %s %s",
		nameStyle.Render("Node: "+node.Name),
		statusIcon(node.Status),
		statusStyle.Render(node.Status)))

	// Dependencies
	if len(node.Dependencies) > 0 {
		lines = append(lines, cli.Muted("Dependencies: ")+strings.Join(node.Dependencies, ", "))
	} else {
		lines = append(lines, cli.Muted("Dependencies: none (entry node)"))
	}

	// Outputs (downstream nodes)
	if len(node.Outputs) > 0 {
		lines = append(lines, cli.Muted("Downstream:   ")+strings.Join(node.Outputs, ", "))
	}

	lines = append(lines, "")

	// Input section (context from dependencies)
	lines = append(lines, cli.Info("Input (from dependencies):"))
	if inputContext != "" && inputContext != "{}" {
		// Format the input nicely
		input := formatJSON(inputContext, 60)
		lines = append(lines, "  "+input)
	} else {
		lines = append(lines, cli.Muted("  (no input data yet)"))
	}

	lines = append(lines, "")

	// Output section (result from this node)
	lines = append(lines, cli.Info("Output (from this node):"))
	if node.Result != "" && node.Result != "{}" {
		// Format the result nicely
		result := formatJSON(node.Result, 60)
		lines = append(lines, "  "+result)
	} else {
		lines = append(lines, cli.Muted("  (no output yet - run node to see results)"))
	}

	return strings.Join(lines, "\n")
}

// RenderNodeCode renders the source code for a node
func (r *GraphRenderer) RenderNodeCode(code *workflow.NodeCodeInfo) string {
	if code == nil {
		return cli.Muted("Source code not available")
	}

	var lines []string
	lines = append(lines, cli.Info(fmt.Sprintf("Source (lines %d-%d):", code.StartLine, code.EndLine)))

	// Show the code with basic formatting
	codeLines := strings.Split(code.Code, "\n")
	maxLines := 8 // Show at most 8 lines of code

	codeStyle := lipgloss.NewStyle().
		Foreground(cli.ColorCode)

	for i, line := range codeLines {
		if i >= maxLines {
			lines = append(lines, cli.Muted(fmt.Sprintf("  ... +%d more lines", len(codeLines)-maxLines)))
			break
		}
		// Truncate long lines
		if len(line) > 70 {
			line = line[:67] + "..."
		}
		lines = append(lines, "  "+codeStyle.Render(line))
	}

	return strings.Join(lines, "\n")
}

// formatJSON formats a JSON string for display, truncating if needed
func formatJSON(jsonStr string, maxLen int) string {
	// Clean up the JSON string for display
	jsonStr = strings.TrimSpace(jsonStr)

	// If it's too long, truncate
	if len(jsonStr) > maxLen {
		jsonStr = jsonStr[:maxLen-3] + "..."
	}

	return jsonStr
}

// RenderHelp renders the help text for the graph view
func (r *GraphRenderer) RenderHelp() string {
	return cli.Muted("←/→: Navigate | Enter: Run | R: Run All | X: Reset | e: Edit | L: Reload | Esc: Back | ?: Help")
}
