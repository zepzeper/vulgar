package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
	openaiSvc "github.com/zepzeper/vulgar/internal/services/openai"
)

const ModuleName = "ai.openai"
const clientTypeName = "openai.client"

type clientWrapper struct {
	service *openaiSvc.Client
}

// luaConfigure configures the OpenAI client
// Usage: local client, err = openai.configure({api_key = "sk-...", org_id = "org-..."})
func luaConfigure(L *lua.LState) int {
	var opts openaiSvc.ClientOptions
	useConfig := true

	if L.GetTop() >= 1 && L.Get(1).Type() == lua.LTTable {
		tbl := L.CheckTable(1)
		useConfig = false // Table provided, use explicit opts

		if val := tbl.RawGetString("api_key"); val.Type() == lua.LTString {
			opts.APIkey = val.String()
		}

		if val := tbl.RawGetString("model"); val.Type() == lua.LTString {
			opts.Model = val.String()
		}

		if val := tbl.RawGetString("temp"); val.Type() == lua.LTString {
			opts.Temp = val.String()
		}
	}

	var client *openaiSvc.Client
	var err error

	if useConfig {
		client, err = openaiSvc.NewClientFromConfig()
	} else {
		client, err = openaiSvc.NewClient(opts)
	}

	if err != nil {
		return util.PushError(L, "failed to configure openai: %v", err)
	}

	ud := L.NewUserData()
	ud.Value = &clientWrapper{service: client}
	L.SetMetatable(ud, L.GetTypeMetatable(clientTypeName))

	L.Push(ud)
	L.Push(lua.LNil)

	return 2
}

func parseChatArgs(L *lua.LState, index int) (openai.ChatCompletionRequest, error) {
	opts := L.CheckTable(index)

	// Sentinel defaults
	model := ""
	var temperature float32 = -1.0

	if m := opts.RawGetString("model"); m.Type() == lua.LTString {
		model = m.String()
	}

	if t := opts.RawGetString("temperature"); t.Type() == lua.LTNumber {
		temperature = float32(t.(lua.LNumber))
	} else if t := opts.RawGetString("temp"); t.Type() == lua.LTNumber {
		temperature = float32(t.(lua.LNumber))
	}

	var messages []openai.ChatCompletionMessage
	if msgs := opts.RawGetString("messages"); msgs.Type() == lua.LTTable {
		msgs.(*lua.LTable).ForEach(func(_, v lua.LValue) {
			if tbl, ok := v.(*lua.LTable); ok {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    tbl.RawGetString("role").String(),
					Content: tbl.RawGetString("content").String(),
				})
			}
		})
	}

	return openai.ChatCompletionRequest{
		Model:       model,
		Temperature: temperature,
		Messages:    messages,
	}, nil
}

// luaChat sends a chat completion request
// Usage: local response, err = openai.chat(client, {model = "gpt-4", messages = {{role = "user", content = "Hello"}}})
func luaChat(L *lua.LState) int {
	ud := L.OptUserData(1, nil)
	if ud == nil {
		return util.PushError(L, "expected openai client")
	}
	wrapper, ok := ud.Value.(*clientWrapper)
	if !ok {
		return util.PushError(L, "expected openai client")
	}

	req, _ := parseChatArgs(L, 2)
	resp, err := wrapper.service.Chat(context.Background(), req)
	if err != nil {
		return util.PushError(L, "chat failed: %v", err)
	}

	result := L.NewTable()
	if len(resp.Choices) > 0 {
		result.RawSetString("content", lua.LString(resp.Choices[0].Message.Content))
		result.RawSetString("role", lua.LString(resp.Choices[0].Message.Role))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaChatStream sends a streaming chat completion request
// Usage: local err = openai.chat_stream(client, {model = "gpt-4", messages = {...}}, function(chunk) print(chunk) end)
func luaChatStream(L *lua.LState) int {
	ud := L.OptUserData(1, nil)
	if ud == nil {
		return util.PushError(L, "expected openai client")
	}
	wrapper, ok := ud.Value.(*clientWrapper)
	if !ok {
		return util.PushError(L, "expected openai client")
	}

	req, _ := parseChatArgs(L, 2)
	callback := L.CheckFunction(3)

	req.Stream = true

	stream, err := wrapper.service.ChatStream(context.Background(), req)
	if err != nil {
		return util.PushError(L, "stream creation failed: %v", err)
	}
	defer stream.Close()

	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}

		if len(resp.Choices) > 0 {
			L.Push(callback)
			L.Push(lua.LString(resp.Choices[0].Delta.Content))
			L.Call(1, 0)
		}
	}

	return 0
}

// luaEmbed creates embeddings
// Usage: local embeddings, err = openai.embeddings(client, {model = "text-embedding-ada-002", input = "Hello world"})
func luaEmbed(L *lua.LState) int {
	ud := L.OptUserData(1, nil)
	if ud == nil {
		return util.PushError(L, "expected openai client")
	}
	wrapper, ok := ud.Value.(*clientWrapper)
	if !ok {
		return util.PushError(L, "expected openai client")
	}

	opts := L.CheckTable(2)
	model := openai.AdaEmbeddingV2
	if m := opts.RawGetString("model"); m.Type() == lua.LTString {
		model = openai.EmbeddingModel(m.String())
	}

	var input []string
	val := opts.RawGetString("input")
	if val.Type() == lua.LTString {
		input = []string{val.String()}
	} else if val.Type() == lua.LTTable {
		val.(*lua.LTable).ForEach(func(_, v lua.LValue) {
			if v.Type() == lua.LTString {
				input = append(input, v.String())
			}
		})
	}

	resp, err := wrapper.service.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Model: model,
		Input: input,
	})
	if err != nil {
		return util.PushError(L, "embeddings failed: %v", err)
	}

	result := L.NewTable()
	for _, data := range resp.Data {
		embedding := L.NewTable()
		for _, v := range data.Embedding {
			embedding.Append(lua.LNumber(v))
		}
		result.Append(embedding)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaImage generates images
// Usage: local images, err = openai.image(client, {prompt = "A cat", size = "1024x1024", n = 1, model = "dall-e-3"})
func luaImage(L *lua.LState) int {
	ud := L.OptUserData(1, nil)
	if ud == nil {
		return util.PushError(L, "expected openai client")
	}
	wrapper, ok := ud.Value.(*clientWrapper)
	if !ok {
		return util.PushError(L, "expected openai client")
	}

	opts := L.CheckTable(2)
	prompt := opts.RawGetString("prompt").String()
	if prompt == "" {
		return util.PushError(L, "prompt is required")
	}

	req := openai.ImageRequest{
		Prompt: prompt,
	}

	if m := opts.RawGetString("model"); m.Type() == lua.LTString {
		req.Model = m.String()
	}
	if s := opts.RawGetString("size"); s.Type() == lua.LTString {
		req.Size = s.String()
	}
	if n := opts.RawGetString("n"); n.Type() == lua.LTNumber {
		req.N = int(n.(lua.LNumber))
	}
	if q := opts.RawGetString("quality"); q.Type() == lua.LTString {
		req.Quality = q.String()
	}
	if style := opts.RawGetString("style"); style.Type() == lua.LTString {
		req.Style = style.String()
	}

	resp, err := wrapper.service.GenerateImage(context.Background(), req)
	if err != nil {
		return util.PushError(L, "image generation failed: %v", err)
	}

	result := L.NewTable()
	for _, data := range resp.Data {
		img := L.NewTable()
		img.RawSetString("url", lua.LString(data.URL))
		img.RawSetString("revised_prompt", lua.LString(data.RevisedPrompt))
		result.Append(img)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaTranscribe transcribes audio
// Usage: local text, err = openai.transcribe(client, {file = "/path/to/audio.mp3", model = "whisper-1"})
func luaTranscribe(L *lua.LState) int {
	ud := L.OptUserData(1, nil)
	if ud == nil {
		return util.PushError(L, "expected openai client")
	}
	wrapper, ok := ud.Value.(*clientWrapper)
	if !ok {
		return util.PushError(L, "expected openai client")
	}

	var filePath string
	var model string

	if L.Get(2).Type() == lua.LTString {
		filePath = L.CheckString(2)
	} else if L.Get(2).Type() == lua.LTTable {
		opts := L.CheckTable(2)
		filePath = opts.RawGetString("file").String()
		if m := opts.RawGetString("model"); m.Type() == lua.LTString {
			model = m.String()
		}
	} else {
		return util.PushError(L, "file path or options table required")
	}

	req := openai.AudioRequest{
		Model:    model,
		FilePath: filePath,
	}

	resp, err := wrapper.service.CreateTranscription(context.Background(), req)
	if err != nil {
		return util.PushError(L, "transcription failed: %v", err)
	}

	L.Push(lua.LString(resp.Text))
	L.Push(lua.LNil)
	return 2
}

// luaTranslate translates audio
// Usage: local text, err = openai.translate(client, {file = "/path/to/audio.mp3", model = "whisper-1"})
func luaTranslate(L *lua.LState) int {
	ud := L.OptUserData(1, nil)
	if ud == nil {
		return util.PushError(L, "expected openai client")
	}
	wrapper, ok := ud.Value.(*clientWrapper)
	if !ok {
		return util.PushError(L, "expected openai client")
	}

	var filePath string
	var model string

	if L.Get(2).Type() == lua.LTString {
		filePath = L.CheckString(2)
	} else if L.Get(2).Type() == lua.LTTable {
		opts := L.CheckTable(2)
		filePath = opts.RawGetString("file").String()
		if m := opts.RawGetString("model"); m.Type() == lua.LTString {
			model = m.String()
		}
	} else {
		return util.PushError(L, "file path or options table required")
	}

	req := openai.AudioRequest{
		Model:    model,
		FilePath: filePath,
	}

	resp, err := wrapper.service.CreateTranslation(context.Background(), req)
	if err != nil {
		return util.PushError(L, "translation failed: %v", err)
	}

	L.Push(lua.LString(resp.Text))
	L.Push(lua.LNil)
	return 2
}

// luaModerate checks content for policy violations
// Usage: local results, err = openai.moderate(client, "Some text to check")
func luaModerate(L *lua.LState) int {
	ud := L.OptUserData(1, nil)
	if ud == nil {
		return util.PushError(L, "expected openai client")
	}
	wrapper, ok := ud.Value.(*clientWrapper)
	if !ok {
		return util.PushError(L, "expected openai client")
	}

	input := L.CheckString(2)
	if input == "" {
		return util.PushError(L, "input text is required")
	}

	resp, err := wrapper.service.Moderation(context.Background(), input)
	if err != nil {
		return util.PushError(L, "moderation failed: %v", err)
	}

	result := L.NewTable()
	for _, res := range resp.Results {
		item := L.NewTable()
		item.RawSetString("flagged", lua.LBool(res.Flagged))

		categories := L.NewTable()
		categories.RawSetString("hate", lua.LBool(res.Categories.Hate))
		categories.RawSetString("hate_threatening", lua.LBool(res.Categories.HateThreatening))
		categories.RawSetString("harassment", lua.LBool(res.Categories.Harassment))
		categories.RawSetString("harassment_threatening", lua.LBool(res.Categories.HarassmentThreatening))
		categories.RawSetString("self_harm", lua.LBool(res.Categories.SelfHarm))
		categories.RawSetString("self_harm_intent", lua.LBool(res.Categories.SelfHarmIntent))
		categories.RawSetString("self_harm_instructions", lua.LBool(res.Categories.SelfHarmInstructions))
		categories.RawSetString("sexual", lua.LBool(res.Categories.Sexual))
		categories.RawSetString("sexual_minors", lua.LBool(res.Categories.SexualMinors))
		categories.RawSetString("violence", lua.LBool(res.Categories.Violence))
		categories.RawSetString("violence_graphic", lua.LBool(res.Categories.ViolenceGraphic))
		item.RawSetString("categories", categories)

		scores := L.NewTable()
		scores.RawSetString("hate", lua.LNumber(res.CategoryScores.Hate))
		scores.RawSetString("hate_threatening", lua.LNumber(res.CategoryScores.Hate))
		scores.RawSetString("harassment", lua.LNumber(res.CategoryScores.Harassment))
		scores.RawSetString("harassment_threatening", lua.LNumber(res.CategoryScores.HarassmentThreatening))
		scores.RawSetString("self_harm", lua.LNumber(res.CategoryScores.SelfHarm))
		scores.RawSetString("self_harm_intent", lua.LNumber(res.CategoryScores.SelfHarmIntent))
		scores.RawSetString("self_harm_instructions", lua.LNumber(res.CategoryScores.SelfHarmInstructions))
		scores.RawSetString("sexual", lua.LNumber(res.CategoryScores.Sexual))
		scores.RawSetString("sexual_minors", lua.LNumber(res.CategoryScores.SexualMinors))
		scores.RawSetString("violence", lua.LNumber(res.CategoryScores.Violence))
		scores.RawSetString("violence_graphic", lua.LNumber(res.CategoryScores.ViolenceGraphic))
		item.RawSetString("category_scores", scores)

		result.Append(item)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// luaListModels lists available models
// Usage: local models, err = openai.list_models(client)
func luaListModels(L *lua.LState) int {
	ud := L.OptUserData(1, nil)
	if ud == nil {
		return util.PushError(L, "expected openai client")
	}
	wrapper, ok := ud.Value.(*clientWrapper)
	if !ok {
		return util.PushError(L, "expected openai client")
	}

	resp, err := wrapper.service.ListModels(context.Background())
	if err != nil {
		return util.PushError(L, "list models failed: %v", err)
	}

	result := L.NewTable()
	for _, m := range resp.Models {
		model := L.NewTable()
		model.RawSetString("id", lua.LString(m.ID))
		model.RawSetString("created", lua.LNumber(m.CreatedAt))
		model.RawSetString("owned_by", lua.LString(m.OwnedBy))
		result.Append(model)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

var exports = map[string]lua.LGFunction{
	"client":      luaConfigure,
	"chat":        luaChat,
	"chat_stream": luaChatStream,
	"embeddings":  luaEmbed,
	"image":       luaImage,
	"transcribe":  luaTranscribe,
	"translate":   luaTranslate,
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
