package nats

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.nats"

// luaConnect connects to a NATS server
// Usage: local client, err = nats.connect({url = "nats://localhost:4222", user = "user", password = "pass"})
func luaConnect(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaPublish publishes a message to a subject
// Usage: local err = nats.publish(client, subject, message)
func luaPublish(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaSubscribe subscribes to a subject
// Usage: local sub, err = nats.subscribe(client, subject, function(msg) print(msg.data) end)
func luaSubscribe(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaQueueSubscribe subscribes to a subject with a queue group
// Usage: local sub, err = nats.queue_subscribe(client, subject, queue, function(msg) print(msg.data) end)
func luaQueueSubscribe(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaRequest sends a request and waits for a response
// Usage: local response, err = nats.request(client, subject, message, timeout)
func luaRequest(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaUnsubscribe unsubscribes from a subscription
// Usage: local err = nats.unsubscribe(sub)
func luaUnsubscribe(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaFlush flushes the connection
// Usage: local err = nats.flush(client)
func luaFlush(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaClose closes the NATS connection
// Usage: local err = nats.close(client)
func luaClose(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"connect":         luaConnect,
	"publish":         luaPublish,
	"subscribe":       luaSubscribe,
	"queue_subscribe": luaQueueSubscribe,
	"request":         luaRequest,
	"unsubscribe":     luaUnsubscribe,
	"flush":           luaFlush,
	"close":           luaClose,
}

// Loader is called when the module is required via require("nats")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
