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

func TestChatStream(t *testing.T) {
	skipIfNoKey(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("api_key", lua.LString(os.Getenv("OPENAI_TEST_API_KEY")))

	err := L.DoString(`
		local openai = require("ai.openai")
		local client, _ = openai.client({api_key = api_key})
		
		local chunks = {}
		local err = openai.chat_stream(client, {
			model = "gpt-3.5-turbo",
			messages = {{role = "user", content = "Count to 3"}}
		}, function(chunk)
			table.insert(chunks, chunk)
		end)
		
		assert(err == nil, "chat_stream should not error: " .. tostring(err))
		assert(#chunks > 0, "should have received chunks")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// embeddings tests
// =============================================================================

func TestEmbeddings(t *testing.T) {
	skipIfNoKey(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("api_key", lua.LString(os.Getenv("OPENAI_TEST_API_KEY")))

	err := L.DoString(`
		local openai = require("ai.openai")
		local client, _ = openai.client({api_key = api_key})
		
		local embeddings, err = openai.embeddings(client, {
			model = "text-embedding-ada-002",
			input = "Hello world"
		})
		
		assert(err == nil, "embeddings should not error: " .. tostring(err))
		assert(embeddings ~= nil, "embeddings should not be nil")
		assert(#embeddings > 0, "should have at least one embedding")
		assert(#embeddings[1] > 0, "embedding vector should not be empty")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// image tests
// =============================================================================

func TestImageGenerate(t *testing.T) {
	skipIfNoKey(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("api_key", lua.LString(os.Getenv("OPENAI_TEST_API_KEY")))

	err := L.DoString(`
		local openai = require("ai.openai")
		local client, _ = openai.client({api_key = api_key})
		
		local images, err = openai.image(client, {
			prompt = "A red circle",
			size = "256x256",
			n = 1
		})
		
		assert(err == nil, "image generation should not error: " .. tostring(err))
		assert(images ~= nil, "images should not be nil")
		assert(#images > 0, "should have at least one image")
		assert(images[1].url ~= nil, "image url should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// transcribe tests
// =============================================================================

func TestTranscribe(t *testing.T) {
	skipIfNoKey(t)
	// TODO: Requires an actual audio file to test properly. skipping for now.
	t.Skip("Skipping transcription integration test (requires audio file)")
}

func TestTranslate(t *testing.T) {
	skipIfNoKey(t)
	// TODO: Requires an actual audio file to test properly. skipping for now.
	t.Skip("Skipping translation integration test (requires audio file)")
}

// =============================================================================
// moderate tests
// =============================================================================

func TestModerate(t *testing.T) {
	skipIfNoKey(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("api_key", lua.LString(os.Getenv("OPENAI_TEST_API_KEY")))

	err := L.DoString(`
		local openai = require("ai.openai")
		local client, _ = openai.client({api_key = api_key})
		
		local results, err = openai.moderate(client, "I want to kill them")
		
		assert(err == nil, "moderation should not error: " .. tostring(err))
		assert(results ~= nil, "results should not be nil")
		assert(#results > 0, "should have results")
		assert(results[1].flagged == true, "should be flagged")
		assert(results[1].categories.violence == true, "should be flagged as violence")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list_models tests
// =============================================================================

func TestListModels(t *testing.T) {
	skipIfNoKey(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("api_key", lua.LString(os.Getenv("OPENAI_TEST_API_KEY")))

	err := L.DoString(`
		local openai = require("ai.openai")
		local client, _ = openai.client({api_key = api_key})
		
		local models, err = openai.list_models(client)
		
		assert(err == nil, "list_models should not error: " .. tostring(err))
		assert(models ~= nil, "models should not be nil")
		assert(#models > 0, "should have at least one model")
		assert(models[1].id ~= nil, "model id should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
