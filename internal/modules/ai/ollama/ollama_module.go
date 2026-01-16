package ollama

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "ai.ollama"

// luaConfigure configures the Ollama client
// Usage: local client, err = ollama.configure({base_url = "http://localhost:11434"})
func luaConfigure(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaGenerate generates a completion
// Usage: local response, err = ollama.generate(client, {model = "llama2", prompt = "Hello"})
func luaGenerate(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaChat sends a chat message
// Usage: local response, err = ollama.chat(client, {model = "llama2", messages = {{role = "user", content = "Hello"}}})
func luaChat(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaChatStream sends a streaming chat message
// Usage: local err = ollama.chat_stream(client, params, function(chunk) print(chunk) end)
func luaChatStream(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaEmbed creates embeddings
// Usage: local embeddings, err = ollama.embed(client, {model = "llama2", prompt = "Hello"})
func luaEmbed(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaListModels lists available models
// Usage: local models, err = ollama.list_models(client)
func luaListModels(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaPullModel pulls a model
// Usage: local err = ollama.pull_model(client, "llama2")
func luaPullModel(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaDeleteModel deletes a model
// Usage: local err = ollama.delete_model(client, "llama2")
func luaDeleteModel(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"configure":    luaConfigure,
	"generate":     luaGenerate,
	"chat":         luaChat,
	"chat_stream":  luaChatStream,
	"embed":        luaEmbed,
	"list_models":  luaListModels,
	"pull_model":   luaPullModel,
	"delete_model": luaDeleteModel,
}

// Loader is called when the module is required via require("ollama")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
