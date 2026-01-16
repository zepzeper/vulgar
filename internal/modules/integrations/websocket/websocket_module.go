package websocket

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.websocket"

// luaConnect connects to a WebSocket server
// Usage: local conn, err = websocket.connect("wss://example.com/ws", {headers = {Authorization = "Bearer ..."}})
func luaConnect(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaSend sends a message over the WebSocket
// Usage: local err = websocket.send(conn, message)
func luaSend(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaSendJSON sends a JSON message over the WebSocket
// Usage: local err = websocket.send_json(conn, {type = "ping"})
func luaSendJSON(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaReceive receives a message from the WebSocket
// Usage: local message, err = websocket.receive(conn)
func luaReceive(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaOnMessage registers a callback for incoming messages
// Usage: local err = websocket.on_message(conn, function(msg) print(msg) end)
func luaOnMessage(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaOnClose registers a callback for connection close
// Usage: local err = websocket.on_close(conn, function() print("closed") end)
func luaOnClose(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaPing sends a ping frame
// Usage: local err = websocket.ping(conn)
func luaPing(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaClose closes the WebSocket connection
// Usage: local err = websocket.close(conn)
func luaClose(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"connect":    luaConnect,
	"send":       luaSend,
	"send_json":  luaSendJSON,
	"receive":    luaReceive,
	"on_message": luaOnMessage,
	"on_close":   luaOnClose,
	"ping":       luaPing,
	"close":      luaClose,
}

// Loader is called when the module is required via require("websocket")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
