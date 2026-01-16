package rabbitmq

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.rabbitmq"

// luaConnect connects to a RabbitMQ server
// Usage: local client, err = rabbitmq.connect({host = "localhost", port = 5672, user = "guest", password = "guest", vhost = "/"})
func luaConnect(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaPublish publishes a message to an exchange
// Usage: local err = rabbitmq.publish(client, exchange, routing_key, message, {content_type = "application/json"})
func luaPublish(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaConsume starts consuming messages from a queue
// Usage: local err = rabbitmq.consume(client, queue, function(msg) print(msg.body) end)
func luaConsume(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaDeclareQueue declares a queue
// Usage: local err = rabbitmq.declare_queue(client, queue, {durable = true, auto_delete = false})
func luaDeclareQueue(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaDeclareExchange declares an exchange
// Usage: local err = rabbitmq.declare_exchange(client, exchange, type, {durable = true})
func luaDeclareExchange(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaBindQueue binds a queue to an exchange
// Usage: local err = rabbitmq.bind_queue(client, queue, exchange, routing_key)
func luaBindQueue(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaAck acknowledges a message
// Usage: local err = rabbitmq.ack(client, delivery_tag)
func luaAck(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaNack negative acknowledges a message
// Usage: local err = rabbitmq.nack(client, delivery_tag, {requeue = true})
func luaNack(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaClose closes the RabbitMQ connection
// Usage: local err = rabbitmq.close(client)
func luaClose(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"connect":          luaConnect,
	"publish":          luaPublish,
	"consume":          luaConsume,
	"declare_queue":    luaDeclareQueue,
	"declare_exchange": luaDeclareExchange,
	"bind_queue":       luaBindQueue,
	"ack":              luaAck,
	"nack":             luaNack,
	"close":            luaClose,
}

// Loader is called when the module is required via require("rabbitmq")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
