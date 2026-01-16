package ftp

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.ftp"

// luaConnect connects to an FTP server
// Usage: local client, err = ftp.connect({host = "example.com", port = 21, user = "user", password = "pass"})
func luaConnect(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaUpload uploads a local file to the remote server
// Usage: local err = ftp.upload(client, local_path, remote_path)
func luaUpload(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaDownload downloads a remote file to local
// Usage: local err = ftp.download(client, remote_path, local_path)
func luaDownload(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaList lists files in a remote directory
// Usage: local files, err = ftp.list(client, remote_path)
func luaList(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaMkdir creates a remote directory
// Usage: local err = ftp.mkdir(client, remote_path)
func luaMkdir(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaRemove removes a remote file
// Usage: local err = ftp.remove(client, remote_path)
func luaRemove(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaRename renames a remote file
// Usage: local err = ftp.rename(client, old_path, new_path)
func luaRename(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaClose closes the FTP connection
// Usage: local err = ftp.close(client)
func luaClose(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"connect":  luaConnect,
	"upload":   luaUpload,
	"download": luaDownload,
	"list":     luaList,
	"mkdir":    luaMkdir,
	"remove":   luaRemove,
	"rename":   luaRename,
	"close":    luaClose,
}

// Loader is called when the module is required via require("ftp")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
