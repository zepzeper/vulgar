package regex

import (
	"regexp"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.regex"

// luaMatch tests if a string matches a pattern
// Usage: local matched = regex.match(pattern, text)
func luaMatch(L *lua.LState) int {
	pattern := L.CheckString(1)
	text := L.CheckString(2)

	re, err := regexp.Compile(pattern)
	if err != nil {
		L.Push(lua.LBool(false))
		return 1
	}

	matched := re.MatchString(text)
	L.Push(lua.LBool(matched))
	return 1
}

// luaFind finds the first match in a string
// Usage: local match, err = regex.find(pattern, text)
func luaFind(L *lua.LState) int {
	pattern := L.CheckString(1)
	text := L.CheckString(2)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return util.PushError(L, "%v", err)
	}

	match := re.FindString(text)
	if match == "" {
		return util.PushSuccess(L, lua.LNil)
	}

	return util.PushSuccess(L, lua.LString(match))
}

// luaFindAll finds all matches in a string
// Usage: local matches, err = regex.find_all(pattern, text)
func luaFindAll(L *lua.LState) int {
	pattern := L.CheckString(1)
	text := L.CheckString(2)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return util.PushError(L, "%v", err)
	}

	matches := re.FindAllString(text, -1)
	tbl := L.NewTable()
	for i, match := range matches {
		tbl.RawSetInt(i+1, lua.LString(match))
	}

	return util.PushSuccess(L, tbl)
}

// luaReplace replaces matches with replacement string
// Usage: local result, err = regex.replace(pattern, text, replacement)
func luaReplace(L *lua.LState) int {
	pattern := L.CheckString(1)
	text := L.CheckString(2)
	replacement := L.CheckString(3)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return util.PushError(L, "%v", err)
	}

	// Use FindStringIndex to find first match and replace manually
	loc := re.FindStringIndex(text)
	if loc == nil {
		return util.PushSuccess(L, lua.LString(text))
	}

	result := text[:loc[0]] + replacement + text[loc[1]:]
	return util.PushSuccess(L, lua.LString(result))
}

// luaReplaceAll replaces all matches with replacement string
// Usage: local result, err = regex.replace_all(pattern, text, replacement)
func luaReplaceAll(L *lua.LState) int {
	pattern := L.CheckString(1)
	text := L.CheckString(2)
	replacement := L.CheckString(3)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return util.PushError(L, "%v", err)
	}

	result := re.ReplaceAllString(text, replacement)
	return util.PushSuccess(L, lua.LString(result))
}

// luaSplit splits a string by pattern
// Usage: local parts, err = regex.split(pattern, text)
func luaSplit(L *lua.LState) int {
	pattern := L.CheckString(1)
	text := L.CheckString(2)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return util.PushError(L, "%v", err)
	}

	parts := re.Split(text, -1)
	tbl := L.NewTable()
	for i, part := range parts {
		tbl.RawSetInt(i+1, lua.LString(part))
	}

	return util.PushSuccess(L, tbl)
}

// luaCapture extracts capture groups from a match
// Usage: local groups, err = regex.capture(pattern, text)
func luaCapture(L *lua.LState) int {
	pattern := L.CheckString(1)
	text := L.CheckString(2)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return util.PushError(L, "%v", err)
	}

	matches := re.FindStringSubmatch(text)
	if len(matches) == 0 {
		return util.PushSuccess(L, lua.LNil)
	}

	// matches[0] is the full match, matches[1:] are the capture groups
	tbl := L.NewTable()
	for i := 1; i < len(matches); i++ {
		tbl.RawSetInt(i, lua.LString(matches[i]))
	}

	return util.PushSuccess(L, tbl)
}

var exports = map[string]lua.LGFunction{
	"match":       luaMatch,
	"find":        luaFind,
	"find_all":    luaFindAll,
	"replace":     luaReplace,
	"replace_all": luaReplaceAll,
	"split":       luaSplit,
	"capture":     luaCapture,
}

// Loader is called when the module is required via require("regex")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
