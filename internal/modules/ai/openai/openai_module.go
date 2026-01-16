package openai

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "ai.openai"

// luaConfigure configures the OpenAI client
// Usage: local client, err = openai.configure({api_key = "sk-...", org_id = "org-..."})
func luaConfigure(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaChat sends a chat completion request
// Usage: local response, err = openai.chat(client, {model = "gpt-4", messages = {{role = "user", content = "Hello"}}})
func luaChat(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaChatStream sends a streaming chat completion request
// Usage: local err = openai.chat_stream(client, {model = "gpt-4", messages = {...}}, function(chunk) print(chunk) end)
func luaChatStream(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaComplete sends a completion request (legacy)
// Usage: local response, err = openai.complete(client, {model = "gpt-3.5-turbo-instruct", prompt = "Hello"})
func luaComplete(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaEmbed creates embeddings
// Usage: local embeddings, err = openai.embed(client, {model = "text-embedding-ada-002", input = "Hello world"})
func luaEmbed(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaImage generates images
// Usage: local images, err = openai.image(client, {prompt = "A cat", size = "1024x1024", n = 1})
func luaImage(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaTranscribe transcribes audio
// Usage: local text, err = openai.transcribe(client, {file = "/path/to/audio.mp3", model = "whisper-1"})
func luaTranscribe(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaModerate checks content for policy violations
// Usage: local results, err = openai.moderate(client, "Some text to check")
func luaModerate(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaListModels lists available models
// Usage: local models, err = openai.list_models(client)
func luaListModels(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

var exports = map[string]lua.LGFunction{
	"configure":   luaConfigure,
	"chat":        luaChat,
	"chat_stream": luaChatStream,
	"complete":    luaComplete,
	"embed":       luaEmbed,
	"image":       luaImage,
	"transcribe":  luaTranscribe,
	"moderate":    luaModerate,
	"list_models": luaListModels,
}

// Loader is called when the module is required via require("openai")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
