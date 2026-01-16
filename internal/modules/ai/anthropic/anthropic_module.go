package anthropic

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "ai.anthropic"

// luaConfigure configures the Anthropic client
// Usage: local client, err = anthropic.configure({api_key = "sk-ant-..."})
func luaConfigure(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaMessage sends a message request (Claude)
// Usage: local response, err = anthropic.message(client, {model = "claude-3-opus-20240229", max_tokens = 1024, messages = {{role = "user", content = "Hello"}}})
func luaMessage(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaMessageStream sends a streaming message request
// Usage: local err = anthropic.message_stream(client, params, function(chunk) print(chunk) end)
func luaMessageStream(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaCountTokens counts tokens in a message
// Usage: local count, err = anthropic.count_tokens(client, {model = "claude-3-opus-20240229", messages = {...}})
func luaCountTokens(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNumber(0))
	L.Push(lua.LString("not implemented"))
	return 2
}

var exports = map[string]lua.LGFunction{
	"configure":      luaConfigure,
	"message":        luaMessage,
	"message_stream": luaMessageStream,
	"count_tokens":   luaCountTokens,
}

// Loader is called when the module is required via require("anthropic")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
