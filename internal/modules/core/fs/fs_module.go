package fs

import (
	"io"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "fs"

// FSConfig holds filesystem configuration
type FSConfig struct {
	BaseDir string
}

const luaFSTypeName = "fs_handle"

// fsHandle represents a configured filesystem handle
type fsHandle struct {
	config FSConfig
}

// registerFSType registers the FS userdata type
func registerFSType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaFSTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), handleMethods))
}

// handleMethods are methods available on fs handle instances
var handleMethods = map[string]lua.LGFunction{
	"read_file":   handleReadFile,
	"write_file":  handleWriteFile,
	"append_file": handleAppendFile,
	"exists":      handleExists,
	"remove":      handleRemove,
	"mkdir":       handleMkdir,
	"list_dir":    handleListDir,
	"copy":        handleCopy,
	"move":        handleMove,
	"stat":        handleStat,
}

// checkFSHandle extracts the fs handle from userdata
func checkFSHandle(L *lua.LState) *fsHandle {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*fsHandle); ok {
		return v
	}
	L.ArgError(1, "fs_handle expected")
	return nil
}

// resolvePath resolves a path relative to the base directory
func (h *fsHandle) resolvePath(path string) string {
	if filepath.IsAbs(path) || h.config.BaseDir == "" {
		return path
	}
	return filepath.Join(h.config.BaseDir, path)
}

// parseConfig extracts configuration from a Lua table
func parseConfig(L *lua.LState, tbl *lua.LTable) FSConfig {
	config := FSConfig{}

	if tbl == nil {
		return config
	}

	if baseDir := L.GetField(tbl, "base_dir"); baseDir != lua.LNil {
		config.BaseDir = lua.LVAsString(baseDir)
	}

	return config
}

// luaNew creates a new FS handle with configuration
// Usage: local myfs = fs.new({ base_dir = "/path/to/data" })
func luaNew(L *lua.LState) int {
	config := parseConfig(L, L.OptTable(1, nil))

	handle := &fsHandle{config: config}

	ud := L.NewUserData()
	ud.Value = handle
	L.SetMetatable(ud, L.GetTypeMetatable(luaFSTypeName))
	L.Push(ud)
	return 1
}

// Handle methods (called with : syntax)
func handleReadFile(L *lua.LState) int {
	h := checkFSHandle(L)
	path := h.resolvePath(L.CheckString(2))
	return readFileImpl(L, path)
}

func handleWriteFile(L *lua.LState) int {
	h := checkFSHandle(L)
	path := h.resolvePath(L.CheckString(2))
	content := L.CheckString(3)
	return writeFileImpl(L, path, content)
}

func handleAppendFile(L *lua.LState) int {
	h := checkFSHandle(L)
	path := h.resolvePath(L.CheckString(2))
	content := L.CheckString(3)
	return appendFileImpl(L, path, content)
}

func handleExists(L *lua.LState) int {
	h := checkFSHandle(L)
	path := h.resolvePath(L.CheckString(2))
	return existsImpl(L, path)
}

func handleRemove(L *lua.LState) int {
	h := checkFSHandle(L)
	path := h.resolvePath(L.CheckString(2))
	return removeImpl(L, path)
}

func handleMkdir(L *lua.LState) int {
	h := checkFSHandle(L)
	path := h.resolvePath(L.CheckString(2))
	return mkdirImpl(L, path)
}

func handleListDir(L *lua.LState) int {
	h := checkFSHandle(L)
	path := h.resolvePath(L.CheckString(2))
	return listDirImpl(L, path)
}

func handleCopy(L *lua.LState) int {
	h := checkFSHandle(L)
	src := h.resolvePath(L.CheckString(2))
	dst := h.resolvePath(L.CheckString(3))
	return copyImpl(L, src, dst)
}

func handleMove(L *lua.LState) int {
	h := checkFSHandle(L)
	src := h.resolvePath(L.CheckString(2))
	dst := h.resolvePath(L.CheckString(3))
	return moveImpl(L, src, dst)
}

func handleStat(L *lua.LState) int {
	h := checkFSHandle(L)
	path := h.resolvePath(L.CheckString(2))
	return statImpl(L, path)
}

// Implementation functions
func readFileImpl(L *lua.LState, path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		return util.PushError(L, "failed to read file: %v", err)
	}
	return util.PushSuccess(L, lua.LString(string(content)))
}

func writeFileImpl(L *lua.LState, path, content string) int {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		L.Push(lua.LString("failed to write file: " + err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func appendFileImpl(L *lua.LState, path, content string) int {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		L.Push(lua.LString("failed to open file for append: " + err.Error()))
		return 1
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		L.Push(lua.LString("failed to append to file: " + err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func existsImpl(L *lua.LState, path string) int {
	_, err := os.Stat(path)
	L.Push(lua.LBool(err == nil))
	return 1
}

func removeImpl(L *lua.LState, path string) int {
	err := os.RemoveAll(path)
	if err != nil {
		L.Push(lua.LString("failed to remove: " + err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func mkdirImpl(L *lua.LState, path string) int {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		L.Push(lua.LString("failed to create directory: " + err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func listDirImpl(L *lua.LState, path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return util.PushError(L, "failed to list directory: %v", err)
	}

	tbl := L.NewTable()
	for i, entry := range entries {
		entryTbl := L.NewTable()
		entryTbl.RawSetString("name", lua.LString(entry.Name()))
		entryTbl.RawSetString("is_dir", lua.LBool(entry.IsDir()))
		tbl.RawSetInt(i+1, entryTbl)
	}
	return util.PushSuccess(L, tbl)
}

func copyImpl(L *lua.LState, src, dst string) int {
	srcFile, err := os.Open(src)
	if err != nil {
		L.Push(lua.LString("failed to open source: " + err.Error()))
		return 1
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		L.Push(lua.LString("failed to create destination: " + err.Error()))
		return 1
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		L.Push(lua.LString("failed to copy: " + err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func moveImpl(L *lua.LState, src, dst string) int {
	err := os.Rename(src, dst)
	if err != nil {
		L.Push(lua.LString("failed to move: " + err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func statImpl(L *lua.LState, path string) int {
	info, err := os.Stat(path)
	if err != nil {
		return util.PushError(L, "failed to stat: %v", err)
	}

	tbl := L.NewTable()
	tbl.RawSetString("name", lua.LString(info.Name()))
	tbl.RawSetString("size", lua.LNumber(info.Size()))
	tbl.RawSetString("is_dir", lua.LBool(info.IsDir()))
	tbl.RawSetString("mod_time", lua.LNumber(info.ModTime().Unix()))
	tbl.RawSetString("mode", lua.LString(info.Mode().String()))

	return util.PushSuccess(L, tbl)
}

// Simple function-based API (no configuration, uses current working directory)
func luaReadFile(L *lua.LState) int {
	return readFileImpl(L, L.CheckString(1))
}

func luaWriteFile(L *lua.LState) int {
	return writeFileImpl(L, L.CheckString(1), L.CheckString(2))
}

func luaAppendFile(L *lua.LState) int {
	return appendFileImpl(L, L.CheckString(1), L.CheckString(2))
}

func luaExists(L *lua.LState) int {
	return existsImpl(L, L.CheckString(1))
}

func luaRemove(L *lua.LState) int {
	return removeImpl(L, L.CheckString(1))
}

func luaMkdir(L *lua.LState) int {
	return mkdirImpl(L, L.CheckString(1))
}

func luaListDir(L *lua.LState) int {
	return listDirImpl(L, L.CheckString(1))
}

func luaCopy(L *lua.LState) int {
	return copyImpl(L, L.CheckString(1), L.CheckString(2))
}

func luaMove(L *lua.LState) int {
	return moveImpl(L, L.CheckString(1), L.CheckString(2))
}

func luaStat(L *lua.LState) int {
	return statImpl(L, L.CheckString(1))
}

// exports defines all functions exposed to Lua
var exports = map[string]lua.LGFunction{
	"new":         luaNew,
	"read_file":   luaReadFile,
	"write_file":  luaWriteFile,
	"append_file": luaAppendFile,
	"exists":      luaExists,
	"remove":      luaRemove,
	"mkdir":       luaMkdir,
	"list_dir":    luaListDir,
	"copy":        luaCopy,
	"move":        luaMove,
	"stat":        luaStat,
}

// Loader is called when the module is required via require("fs")
func Loader(L *lua.LState) int {
	registerFSType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
