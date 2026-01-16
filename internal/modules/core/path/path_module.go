package path

import (
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "path"

func luaJoin(L *lua.LState) int {
	n := L.GetTop()

	parts := make([]string, n)
	for i := 1; i <= n; i++ {
		parts[i-1] = L.CheckString(i)
	}

	L.Push(lua.LString(filepath.Join(parts...)))
	return 1
}

func luaDir(L *lua.LState) int {
	p := L.CheckString(1)
	L.Push(lua.LString(filepath.Dir(p)))
	return 1
}

func luaBase(L *lua.LState) int {
	p := L.CheckString(1)
	L.Push(lua.LString(filepath.Base(p)))
	return 1
}

func luaExt(L *lua.LState) int {
	p := L.CheckString(1)
	L.Push(lua.LString(filepath.Ext(p)))
	return 1
}

func luaAbs(L *lua.LState) int {
	p := L.CheckString(1)

	abs, err := filepath.Abs(p)
	if err != nil {
		return util.PushError(L, "%v", err)
	}

	return util.PushSuccess(L, lua.LString(abs))
}

func luaClean(L *lua.LState) int {
	p := L.CheckString(1)
	L.Push(lua.LString(filepath.Clean(p)))
	return 1
}

func luaSplit(L *lua.LState) int {
	p := L.CheckString(1)
	dir, file := filepath.Split(p)
	L.Push(lua.LString(dir))
	L.Push(lua.LString(file))
	return 2
}

func luaIsAbs(L *lua.LState) int {
	p := L.CheckString(1)
	L.Push(lua.LBool(filepath.IsAbs(p)))
	return 1
}

var exports = map[string]lua.LGFunction{
	"join":   luaJoin,
	"dir":    luaDir,
	"base":   luaBase,
	"ext":    luaExt,
	"abs":    luaAbs,
	"clean":  luaClean,
	"split":  luaSplit,
	"is_abs": luaIsAbs,
}

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)

	// Add separator constant
	mod.RawSetString("separator", lua.LString(string(filepath.Separator)))

	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
