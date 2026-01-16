package openai

import (
	"os"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

func skipIfNoKey(t *testing.T) {
	if os.Getenv("OPENAI_TEST_API_KEY") == "" {
		t.Skip("OPENAI_TEST_API_KEY not set, skipping integration test")
	}
}

// =============================================================================
// client tests
// =============================================================================

func TestClientCreate(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local openai = require("ai.openai")
		local client, err = openai.client({api_key = "test-key"})
		assert(err == nil, "client should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestClientMissingKey(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local openai = require("ai.openai")
		local client, err = openai.client({})
		assert(client == nil or err ~= nil, "should error with missing key")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// chat tests
// =============================================================================

func TestChatNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local openai = require("ai.openai")
		local response, err = openai.chat(nil, {
			model = "gpt-3.5-turbo",
			messages = {{role = "user", content = "Hello"}}
		})
		assert(response == nil, "response should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestChat(t *testing.T) {
	skipIfNoKey(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("api_key", lua.LString(os.Getenv("OPENAI_TEST_API_KEY")))

	err := L.DoString(`
		local openai = require("ai.openai")
		local client, _ = openai.client({api_key = api_key})
		
		local response, err = openai.chat(client, {
			model = "gpt-3.5-turbo",
			messages = {{role = "user", content = "Say hi"}}
		})
		assert(err == nil, "chat should not error: " .. tostring(err))
		assert(response ~= nil, "response should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// complete tests
// =============================================================================

func TestCompleteNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local openai = require("ai.openai")
		local response, err = openai.complete(nil, {
			model = "text-davinci-003",
			prompt = "Hello"
		})
		assert(response == nil, "response should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// embeddings tests
// =============================================================================

func TestEmbeddingsNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local openai = require("ai.openai")
		local embeddings, err = openai.embeddings(nil, {
			model = "text-embedding-ada-002",
			input = "Hello world"
		})
		assert(embeddings == nil, "embeddings should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// image tests
// =============================================================================

func TestImageGenerateNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local openai = require("ai.openai")
		local images, err = openai.image(nil, {
			prompt = "A cat",
			n = 1
		})
		assert(images == nil, "images should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// transcribe tests
// =============================================================================

func TestTranscribeNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local openai = require("ai.openai")
		local text, err = openai.transcribe(nil, "/path/to/audio.mp3")
		assert(text == nil, "text should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// translate tests
// =============================================================================

func TestTranslateNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local openai = require("ai.openai")
		local text, err = openai.translate(nil, "/path/to/audio.mp3")
		assert(text == nil, "text should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// moderate tests
// =============================================================================

func TestModerateNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local openai = require("ai.openai")
		local result, err = openai.moderate(nil, "Some text to check")
		assert(result == nil, "result should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list_models tests
// =============================================================================

func TestListModelsNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local openai = require("ai.openai")
		local models, err = openai.list_models(nil)
		assert(models == nil, "models should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
