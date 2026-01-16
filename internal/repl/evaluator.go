package repl

import (
	"strings"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/engine"
)

// Evaluate executes Lua code and returns the results
// It tries to evaluate as an expression first (to capture return values),
// then falls back to statement execution
func Evaluate(eng *engine.Engine, code string) ([]lua.LValue, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil
	}

	L := eng.L

	// Try as expression first (wrap in return statement to capture value)
	if !isStatement(code) {
		exprCode := "return " + code
		if fn, err := L.LoadString(exprCode); err == nil {
			L.Push(fn)
			if err := L.PCall(0, lua.MultRet, nil); err == nil {
				// Collect results from stack
				results := collectResults(L)
				if len(results) > 0 {
					return results, nil
				}
			}
			// If expression evaluation failed or returned nothing,
			// fall through to try as statement
		}
	}

	// Execute as statement
	if err := eng.Eval(code); err != nil {
		return nil, err
	}

	return nil, nil
}

// collectResults pops all values from the Lua stack
func collectResults(L *lua.LState) []lua.LValue {
	n := L.GetTop()
	if n == 0 {
		return nil
	}

	results := make([]lua.LValue, n)
	for i := 1; i <= n; i++ {
		results[i-1] = L.Get(i)
	}
	L.SetTop(0)

	// Filter out nil values at the end
	for len(results) > 0 && results[len(results)-1] == lua.LNil {
		results = results[:len(results)-1]
	}

	return results
}

// isStatement checks if code is clearly a statement (not an expression)
func isStatement(code string) bool {
	trimmed := strings.TrimSpace(code)

	// Keywords that start statements
	statementStarters := []string{
		"local ", "function ", "for ", "while ", "repeat ",
		"if ", "do ", "return ", "break", "goto ",
	}

	for _, starter := range statementStarters {
		if strings.HasPrefix(trimmed, starter) {
			return true
		}
	}

	// Check for assignment (but not ==, ~=, <=, >=)
	if isAssignment(code) {
		return true
	}

	return false
}

// isAssignment checks if code contains an assignment operator
func isAssignment(code string) bool {
	// Find = that isn't part of ==, ~=, <=, >=
	for i := 0; i < len(code); i++ {
		if code[i] == '=' {
			// Check previous char
			if i > 0 {
				prev := code[i-1]
				if prev == '=' || prev == '~' || prev == '<' || prev == '>' {
					continue
				}
			}
			// Check next char
			if i+1 < len(code) && code[i+1] == '=' {
				continue
			}
			return true
		}
	}
	return false
}
