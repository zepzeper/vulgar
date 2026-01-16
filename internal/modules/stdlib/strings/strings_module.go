package strings

import (
	"regexp"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
)

const ModuleName = "stdlib.strings"

// luaTrim trims whitespace from both ends
// Usage: local trimmed = strings.trim("  hello  ")
func luaTrim(L *lua.LState) int {
	name := L.CheckString(1)
	L.Push(lua.LString(strings.TrimSpace(name)))
	return 1
}

// luaTrimLeft trims whitespace from left
// Usage: local trimmed = strings.trim_left("  hello")
func luaTrimLeft(L *lua.LState) int {
	name := L.CheckString(1)
	L.Push(lua.LString(strings.TrimLeft(name, " \t\n\r")))
	return 1
}

// luaTrimRight trims whitespace from right
// Usage: local trimmed = strings.trim_right("hello  ")
func luaTrimRight(L *lua.LState) int {
	str := L.CheckString(1)
	L.Push(lua.LString(strings.TrimRight(str, " \t\n\r")))
	return 1
}

// luaSplit splits a string by separator
// Usage: local parts = strings.split("a,b,c", ",")
func luaSplit(L *lua.LState) int {
	str := L.CheckString(1)
	separator := L.CheckString(2)

	stringSlices := strings.Split(str, separator)

	tbl := L.NewTable()
	for i, part := range stringSlices {
		tbl.RawSetInt(i+1, lua.LString(part)) // Lua arrays are 1-indexed
	}

	L.Push(tbl)
	return 1
}

// luaJoin joins array elements with separator
// Usage: local str = strings.join({"a", "b", "c"}, ",")
func luaJoin(L *lua.LState) int {
	tbl := L.CheckTable(1)
	separator := L.CheckString(2)

	var parts []string
	tbl.ForEach(func(key, value lua.LValue) {
		if str, ok := value.(lua.LString); ok {
			parts = append(parts, string(str))
		}
	})

	L.Push(lua.LString(strings.Join(parts, separator)))
	return 1
}

// luaReplace replaces occurrences of a substring
// Usage: local result = strings.replace("hello world", "world", "lua")
// Usage: local result = strings.replace("hello world", "world", "lua", 2)  -- optional count
func luaReplace(L *lua.LState) int {
	str := L.CheckString(1)
	target := L.CheckString(2)
	replacement := L.CheckString(3)

	n := L.OptInt(4, 1)

	L.Push(lua.LString(strings.Replace(str, target, replacement, n)))
	return 1
}

// luaReplaceAll replaces all occurrences
// Usage: local result = strings.replace_all("aaa", "a", "b")
func luaReplaceAll(L *lua.LState) int {
	str := L.CheckString(1)
	target := L.CheckString(2)
	replacement := L.CheckString(3)
	L.Push(lua.LString(strings.ReplaceAll(str, target, replacement)))
	return 1
}

// luaToUpper converts to uppercase
// Usage: local upper = strings.to_upper("hello")
func luaToUpper(L *lua.LState) int {
	str := L.CheckString(1)
	L.Push(lua.LString(strings.ToUpper(str)))
	return 1
}

// luaToLower converts to lowercase
// Usage: local lower = strings.to_lower("HELLO")
func luaToLower(L *lua.LState) int {
	str := L.CheckString(1)
	L.Push(lua.LString(strings.ToLower(str)))
	return 1
}

// luaCapitalize capitalizes first letter
// Usage: local cap = strings.capitalize("hello")
func luaCapitalize(L *lua.LState) int {
	str := L.CheckString(1)
	if len(str) == 0 {
		L.Push(lua.LString(""))
		return 1
	}
	// Capitalize first letter, lowercase the rest
	result := strings.ToUpper(string(str[0])) + strings.ToLower(str[1:])
	L.Push(lua.LString(result))
	return 1
}

// luaTitle converts to title case
// Usage: local title = strings.title("hello world")
func luaTitle(L *lua.LState) int {
	str := L.CheckString(1)
	// Simple title case implementation (capitalize first letter of each word)
	result := strings.ToLower(str)
	words := strings.Fields(result)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	L.Push(lua.LString(strings.Join(words, " ")))
	return 1
}

// luaSnakeCase converts to snake_case
// Usage: local snake = strings.snake_case("helloWorld")
func luaSnakeCase(L *lua.LState) int {
	str := L.CheckString(1)
	// Convert camelCase/PascalCase to snake_case
	// This is simplified - you may want a more robust implementation
	reg := regexp.MustCompile(`([a-z])([A-Z])`)
	str = reg.ReplaceAllString(str, "${1}_${2}")
	L.Push(lua.LString(strings.ToLower(str)))
	return 1
}

// luaCamelCase converts to camelCase
// Usage: local camel = strings.camel_case("hello_world")
func luaCamelCase(L *lua.LState) int {
	str := L.CheckString(1)
	parts := strings.Split(str, "_")
	if len(parts) == 0 {
		L.Push(lua.LString(""))
		return 1
	}

	result := strings.ToLower(parts[0])
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(string(parts[i][0])) + strings.ToLower(parts[i][1:])
		}
	}
	L.Push(lua.LString(result))
	return 1
}

// luaKebabCase converts to kebab-case
// Usage: local kebab = strings.kebab_case("hello_world")
func luaKebabCase(L *lua.LState) int {
	str := L.CheckString(1)

	// Replace underscores with hyphens
	str = strings.ReplaceAll(str, "_", "-")

	// Handle camelCase/PascalCase by inserting hyphens before uppercase letters
	reg := regexp.MustCompile(`([a-z])([A-Z])`)
	str = reg.ReplaceAllString(str, "${1}-${2}")

	// Convert to lowercase
	str = strings.ToLower(str)

	L.Push(lua.LString(str))
	return 1
}

// luaPascalCase converts to PascalCase
// Usage: local pascal = strings.pascal_case("hello_world")
func luaPascalCase(L *lua.LState) int {
	str := L.CheckString(1)
	parts := strings.Split(str, "_")
	var result strings.Builder

	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(string(part[0])))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}
	L.Push(lua.LString(result.String()))
	return 1
}

// luaContains checks if string contains substring
// Usage: local found = strings.contains("hello world", "world")
func luaContains(L *lua.LState) int {
	str := L.CheckString(1)
	substr := L.CheckString(2)
	L.Push(lua.LBool(strings.Contains(str, substr)))
	return 1
}

// luaStartsWith checks if string starts with prefix
// Usage: local starts = strings.starts_with("hello", "he")
func luaStartsWith(L *lua.LState) int {
	str := L.CheckString(1)
	prefix := L.CheckString(2)
	L.Push(lua.LBool(strings.HasPrefix(str, prefix)))
	return 1
}

// luaEndsWith checks if string ends with suffix
// Usage: local ends = strings.ends_with("hello", "lo")
func luaEndsWith(L *lua.LState) int {
	str := L.CheckString(1)
	suffix := L.CheckString(2)
	L.Push(lua.LBool(strings.HasSuffix(str, suffix)))
	return 1
}

// luaPadLeft pads string on the left
// Usage: local padded = strings.pad_left("42", 5, "0")
func luaPadLeft(L *lua.LState) int {
	str := L.CheckString(1)
	length := L.CheckInt(2)
	padChar := L.OptString(3, " ") // Default to space

	if len(padChar) == 0 {
		padChar = " "
	}

	currentLen := len(str)
	if currentLen >= length {
		L.Push(lua.LString(str))
		return 1
	}

	padCount := length - currentLen
	padding := strings.Repeat(padChar, padCount)
	L.Push(lua.LString(padding + str))
	return 1
}

// luaPadRight pads string on the right
// Usage: local padded = strings.pad_right("hi", 5)
func luaPadRight(L *lua.LState) int {
	str := L.CheckString(1)
	length := L.CheckInt(2)
	padChar := L.OptString(3, " ") // Default to space

	if len(padChar) == 0 {
		padChar = " "
	}

	currentLen := len(str)
	if currentLen >= length {
		L.Push(lua.LString(str))
		return 1
	}

	padCount := length - currentLen
	padding := strings.Repeat(padChar, padCount)
	L.Push(lua.LString(str + padding))
	return 1
}

// luaReverse reverses a string
// Usage: local reversed = strings.reverse("hello")
func luaReverse(L *lua.LState) int {
	str := L.CheckString(1)
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	L.Push(lua.LString(string(runes)))
	return 1
}

// luaTruncate truncates string to length with optional suffix
// Usage: local truncated = strings.truncate("hello world", 8, "...")
func luaTruncate(L *lua.LState) int {
	str := L.CheckString(1)
	maxLen := L.CheckInt(2)
	suffix := L.OptString(3, "...")

	if len(str) <= maxLen {
		L.Push(lua.LString(str))
		return 1
	}

	// Account for suffix length
	truncateLen := max(maxLen-len(suffix), 0)

	result := str[:truncateLen] + suffix
	L.Push(lua.LString(result))
	return 1
}

// luaSlugify converts to URL-safe slug
// Usage: local slug = strings.slugify("Hello World!")
func luaSlugify(L *lua.LState) int {
	str := L.CheckString(1)

	// Convert to lowercase
	str = strings.ToLower(str)

	// Replace spaces and special chars with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	str = reg.ReplaceAllString(str, "-")

	// Remove leading/trailing hyphens
	str = strings.Trim(str, "-")

	L.Push(lua.LString(str))
	return 1
}

var exports = map[string]lua.LGFunction{
	"trim":        luaTrim,
	"trim_left":   luaTrimLeft,
	"trim_right":  luaTrimRight,
	"split":       luaSplit,
	"join":        luaJoin,
	"replace":     luaReplace,
	"replace_all": luaReplaceAll,
	"to_upper":    luaToUpper,
	"to_lower":    luaToLower,
	"capitalize":  luaCapitalize,
	"title":       luaTitle,
	"snake_case":  luaSnakeCase,
	"camel_case":  luaCamelCase,
	"kebab_case":  luaKebabCase,
	"pascal_case": luaPascalCase,
	"contains":    luaContains,
	"starts_with": luaStartsWith,
	"ends_with":   luaEndsWith,
	"pad_left":    luaPadLeft,
	"pad_right":   luaPadRight,
	"reverse":     luaReverse,
	"truncate":    luaTruncate,
	"slugify":     luaSlugify,
}

// Loader is called when the module is required via require("strings")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
