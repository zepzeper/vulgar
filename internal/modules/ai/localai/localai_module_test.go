package localai

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

func skipIfNoLocalAI(t *testing.T) {
	if os.Getenv("LOCALAI_TEST_HOST") == "" {
		t.Skip("LOCALAI_TEST_HOST not set, skipping integration test")
	}
}

// =============================================================================
// client tests
// =============================================================================

func TestClientDefault(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local localai = require("ai.localai")
		local client, err = localai.client()
		assert(err == nil, "client should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestClientCustomHost(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local localai = require("ai.localai")
		local client, err = localai.client({host = "http://localhost:8080"})
		assert(err == nil, "client should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
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
		local localai = require("ai.localai")
		local response, err = localai.chat(nil, {
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

// =============================================================================
// complete tests
// =============================================================================

func TestCompleteNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local localai = require("ai.localai")
		local response, err = localai.complete(nil, {
			model = "gpt-3.5-turbo",
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
		local localai = require("ai.localai")
		local embeddings, err = localai.embeddings(nil, {
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
		local localai = require("ai.localai")
		local images, err = localai.image(nil, {
			prompt = "A cat"
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
		local localai = require("ai.localai")
		local text, err = localai.transcribe(nil, "/path/to/audio.mp3")
		assert(text == nil, "text should be nil")
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
		local localai = require("ai.localai")
		local models, err = localai.list_models(nil)
		assert(models == nil, "models should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// health tests
// =============================================================================

func TestHealthNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local localai = require("ai.localai")
		local health, err = localai.health(nil)
		assert(health == nil, "health should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestHealth(t *testing.T) {
	skipIfNoLocalAI(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("host", lua.LString(os.Getenv("LOCALAI_TEST_HOST")))

	err := L.DoString(`
		local localai = require("ai.localai")
		local client, _ = localai.client({host = host})
		
		local health, err = localai.health(client)
		assert(err == nil, "health should not error: " .. tostring(err))
		assert(health ~= nil, "health should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
