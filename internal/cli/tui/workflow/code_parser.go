package workflow

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// NodeCodeInfo contains the source code location and content for a node
type NodeCodeInfo struct {
	Name      string
	StartLine int
	EndLine   int
	Code      string
}

// ParseNodeCode parses a Lua workflow file and extracts code information for each node
func ParseNodeCode(filePath string) (map[string]NodeCodeInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	nodes := make(map[string]NodeCodeInfo)

	// Read all lines
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Pattern to find workflow.node calls
	// Matches: workflow.node(wf, "name", function(ctx) or workflow.node(wf, 'name', function(ctx)
	nodePattern := regexp.MustCompile(`workflow\.node\s*\(\s*\w+\s*,\s*["']([^"']+)["']\s*,\s*function`)

	// Track nesting for end detection
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		matches := nodePattern.FindStringSubmatch(line)
		if len(matches) >= 2 {
			nodeName := matches[1]
			startLine := i + 1 // 1-indexed

			// Find the matching end by counting function/end pairs
			endLine := findNodeEnd(lines, i)

			// Extract the code
			var codeLines []string
			for j := i; j < endLine && j < len(lines); j++ {
				codeLines = append(codeLines, lines[j])
			}

			nodes[nodeName] = NodeCodeInfo{
				Name:      nodeName,
				StartLine: startLine,
				EndLine:   endLine,
				Code:      strings.Join(codeLines, "\n"),
			}
		}
	}

	return nodes, nil
}

// findNodeEnd finds the end line of a node definition by tracking nesting
func findNodeEnd(lines []string, startIdx int) int {
	depth := 0
	inNode := false

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Count function openings (including the initial workflow.node function)
		if strings.Contains(line, "function") {
			depth++
			inNode = true
		}

		// Count ends - but be careful about "end)" which closes both function and node call
		if inNode {
			// Check for standalone "end" or "end)" or "end, {" etc.
			if strings.HasPrefix(trimmed, "end") {
				depth--
				if depth <= 0 {
					return i + 1 // 1-indexed
				}
			}
		}
	}

	// If we didn't find a proper end, return the end of file
	return len(lines)
}

// GetNodeLineNumber returns the starting line number for a specific node
func GetNodeLineNumber(filePath, nodeName string) (int, error) {
	nodes, err := ParseNodeCode(filePath)
	if err != nil {
		return 0, err
	}

	if info, exists := nodes[nodeName]; exists {
		return info.StartLine, nil
	}

	return 0, nil
}
