package localai

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "ai.localai"

// luaConfigure configures the LocalAI client
// Usage: local client, err = localai.configure({base_url = "http://localhost:8080"})
func luaConfigure(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaChat sends a chat completion request
// Usage: local response, err = localai.chat(client, {model = "gpt4all", messages = {{role = "user", content = "Hello"}}})
func luaChat(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaChatStream sends a streaming chat completion request
// Usage: local err = localai.chat_stream(client, {model = "gpt4all", messages = {...}}, function(chunk) print(chunk) end)
func luaChatStream(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaComplete sends a completion request
// Usage: local response, err = localai.complete(client, {model = "gpt4all", prompt = "Hello"})
func luaComplete(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaEmbed creates embeddings
// Usage: local embeddings, err = localai.embed(client, {model = "all-minilm", input = "Hello world"})
func luaEmbed(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaTranscribe transcribes audio
// Usage: local text, err = localai.transcribe(client, {file = "/path/to/audio.mp3", model = "whisper"})
func luaTranscribe(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaTTS generates speech from text
// Usage: local audio, err = localai.tts(client, {model = "tts", input = "Hello world"})
func luaTTS(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaListModels lists available models
// Usage: local models, err = localai.list_models(client)
func luaListModels(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaHealth checks the LocalAI server health
// Usage: local ok, err = localai.health(client)
func luaHealth(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LBool(false))
	L.Push(lua.LString("not implemented"))
	return 2
}

var exports = map[string]lua.LGFunction{
	"configure":   luaConfigure,
	"chat":        luaChat,
	"chat_stream": luaChatStream,
	"complete":    luaComplete,
	"embed":       luaEmbed,
	"transcribe":  luaTranscribe,
	"tts":         luaTTS,
	"list_models": luaListModels,
	"health":      luaHealth,
}

// Loader is called when the module is required via require("localai")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
