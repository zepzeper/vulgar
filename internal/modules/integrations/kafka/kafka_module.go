package kafka

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.kafka"

// luaConnect connects to a Kafka cluster
// Usage: local client, err = kafka.connect({brokers = {"localhost:9092"}, group_id = "my-group"})
func luaConnect(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaProduce sends a message to a topic
// Usage: local err = kafka.produce(client, topic, message, {key = "key", partition = 0})
func luaProduce(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaConsume starts consuming messages from topics
// Usage: local err = kafka.consume(client, {"topic1", "topic2"}, function(msg) print(msg.value) end)
func luaConsume(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaSubscribe subscribes to topics
// Usage: local err = kafka.subscribe(client, {"topic1", "topic2"})
func luaSubscribe(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaCommit commits consumed offsets
// Usage: local err = kafka.commit(client)
func luaCommit(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaCreateTopic creates a new topic
// Usage: local err = kafka.create_topic(client, topic, {partitions = 3, replication = 1})
func luaCreateTopic(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaDeleteTopic deletes a topic
// Usage: local err = kafka.delete_topic(client, topic)
func luaDeleteTopic(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaListTopics lists all topics
// Usage: local topics, err = kafka.list_topics(client)
func luaListTopics(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaClose closes the Kafka connection
// Usage: local err = kafka.close(client)
func luaClose(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"connect":      luaConnect,
	"produce":      luaProduce,
	"consume":      luaConsume,
	"subscribe":    luaSubscribe,
	"commit":       luaCommit,
	"create_topic": luaCreateTopic,
	"delete_topic": luaDeleteTopic,
	"list_topics":  luaListTopics,
	"close":        luaClose,
}

// Loader is called when the module is required via require("kafka")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
