package repl

import (
	"fmt"
	"sort"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

const (
	maxDepth      = 10
	maxArrayItems = 100
	maxTableItems = 50
	maxStringLen  = 500
	indentString  = "  "
)

// PrintValues formats and prints Lua values to stdout
func PrintValues(values []lua.LValue) {
	if len(values) == 0 {
		return
	}

	for i, v := range values {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(FormatValue(v, 0))
	}
	fmt.Println()
}

// FormatValue converts a Lua value to a formatted string
func FormatValue(v lua.LValue, depth int) string {
	if depth > maxDepth {
		return "..."
	}

	switch val := v.(type) {
	case *lua.LNilType:
		return nilStyle.Render("nil")

	case lua.LBool:
		if val {
			return successStyle.Render("true")
		}
		return errorStyle.Render("false")

	case lua.LNumber:
		return numberStyle.Render(fmt.Sprintf("%v", float64(val)))

	case lua.LString:
		s := string(val)
		if len(s) > maxStringLen {
			s = s[:maxStringLen] + "..."
		}
		return stringStyle.Render(fmt.Sprintf("%q", s))

	case *lua.LTable:
		return formatTable(val, depth)

	case *lua.LFunction:
		return nilStyle.Render("<function>")

	case *lua.LUserData:
		return nilStyle.Render("<userdata>")

	case *lua.LState:
		return nilStyle.Render("<thread>")

	case *lua.LChannel:
		return nilStyle.Render("<channel>")

	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatTable formats a Lua table with proper indentation
func formatTable(t *lua.LTable, depth int) string {
	if depth > maxDepth {
		return "{...}"
	}

	// Collect all keys
	var arrayPart []lua.LValue
	var hashPart []struct {
		key   lua.LValue
		value lua.LValue
	}

	maxArrayIndex := 0
	t.ForEach(func(k, v lua.LValue) {
		if num, ok := k.(lua.LNumber); ok {
			idx := int(num)
			if float64(idx) == float64(num) && idx >= 1 {
				if idx > maxArrayIndex {
					maxArrayIndex = idx
				}
			}
		}
	})

	// Check if it's a proper array (consecutive indices from 1)
	isArray := maxArrayIndex > 0
	if isArray {
		for i := 1; i <= maxArrayIndex; i++ {
			if t.RawGetInt(i) == lua.LNil {
				isArray = false
				break
			}
		}
	}

	// Separate array and hash parts
	t.ForEach(func(k, v lua.LValue) {
		if isArray {
			if num, ok := k.(lua.LNumber); ok {
				idx := int(num)
				if idx >= 1 && idx <= maxArrayIndex {
					// Will add in order below
					return
				}
			}
		}
		hashPart = append(hashPart, struct {
			key   lua.LValue
			value lua.LValue
		}{k, v})
	})

	// Get array values in order
	if isArray {
		for i := 1; i <= maxArrayIndex && i <= maxArrayItems; i++ {
			arrayPart = append(arrayPart, t.RawGetInt(i))
		}
	}

	// Sort hash keys for consistent output
	sort.Slice(hashPart, func(i, j int) bool {
		return formatKey(hashPart[i].key) < formatKey(hashPart[j].key)
	})

	// Limit items
	if len(hashPart) > maxTableItems {
		hashPart = hashPart[:maxTableItems]
	}

	// Build output
	var parts []string
	indent := strings.Repeat(indentString, depth+1)

	// Add array items
	for i, v := range arrayPart {
		if i >= maxArrayItems {
			parts = append(parts, "...")
			break
		}
		parts = append(parts, FormatValue(v, depth+1))
	}

	// Add hash items
	truncatedHash := len(hashPart) > maxTableItems
	for i, kv := range hashPart {
		if i >= maxTableItems {
			break
		}
		keyStr := formatKey(kv.key)
		valueStr := FormatValue(kv.value, depth+1)
		parts = append(parts, keyStyle.Render(keyStr)+" = "+valueStr)
	}
	if truncatedHash {
		parts = append(parts, "...")
	}

	// Format output
	if len(parts) == 0 {
		return "{}"
	}

	// Use compact format for small tables
	totalLen := 0
	hasNested := false
	for _, p := range parts {
		totalLen += len(p)
		if strings.Contains(p, "\n") {
			hasNested = true
		}
	}

	if len(parts) <= 5 && totalLen < 60 && !hasNested {
		return "{ " + strings.Join(parts, ", ") + " }"
	}

	// Multi-line format
	var sb strings.Builder
	sb.WriteString("{\n")
	for _, p := range parts {
		sb.WriteString(indent)
		sb.WriteString(p)
		sb.WriteString(",\n")
	}
	sb.WriteString(strings.Repeat(indentString, depth))
	sb.WriteString("}")
	return sb.String()
}

// formatKey formats a table key for display
func formatKey(k lua.LValue) string {
	switch key := k.(type) {
	case lua.LString:
		s := string(key)
		// Check if it's a valid identifier
		if isValidIdentifier(s) {
			return s
		}
		return fmt.Sprintf("[%q]", s)
	case lua.LNumber:
		return fmt.Sprintf("[%v]", float64(key))
	default:
		return fmt.Sprintf("[%v]", k)
	}
}

// isValidIdentifier checks if a string is a valid Lua identifier
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_') {
				return false
			}
		} else {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
				return false
			}
		}
	}
	return true
}
