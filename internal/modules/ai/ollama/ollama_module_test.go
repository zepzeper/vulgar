package ollama

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

func skipIfNoOllama(t *testing.T) {
	if os.Getenv("OLLAMA_TEST_HOST") == "" {
		t.Skip("OLLAMA_TEST_HOST not set, skipping integration test")
	}
}

// =============================================================================
// client tests
// =============================================================================

func TestClientDefault(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ollama = require("ai.ollama")
		local client, err = ollama.client()
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
		local ollama = require("ai.ollama")
		local client, err = ollama.client({host = "http://localhost:11434"})
		assert(err == nil, "client should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// generate tests
// =============================================================================

func TestGenerateNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ollama = require("ai.ollama")
		local response, err = ollama.generate(nil, "llama2", "Hello")
		assert(response == nil, "response should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestGenerate(t *testing.T) {
	skipIfNoOllama(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("host", lua.LString(os.Getenv("OLLAMA_TEST_HOST")))
	L.SetGlobal("model", lua.LString(os.Getenv("OLLAMA_TEST_MODEL")))

	err := L.DoString(`
		local ollama = require("ai.ollama")
		local client, _ = ollama.client({host = host})
		
		local response, err = ollama.generate(client, model, "Say hello")
		assert(err == nil, "generate should not error: " .. tostring(err))
		assert(response ~= nil, "response should not be nil")
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
		local ollama = require("ai.ollama")
		local response, err = ollama.chat(nil, "llama2", {
			{role = "user", content = "Hello"}
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
		local ollama = require("ai.ollama")
		local embeddings, err = ollama.embeddings(nil, "llama2", "Hello world")
		assert(embeddings == nil, "embeddings should be nil")
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
		local ollama = require("ai.ollama")
		local models, err = ollama.list_models(nil)
		assert(models == nil, "models should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestListModels(t *testing.T) {
	skipIfNoOllama(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("host", lua.LString(os.Getenv("OLLAMA_TEST_HOST")))

	err := L.DoString(`
		local ollama = require("ai.ollama")
		local client, _ = ollama.client({host = host})
		
		local models, err = ollama.list_models(client)
		assert(err == nil, "list_models should not error: " .. tostring(err))
		assert(models ~= nil, "models should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// pull_model tests
// =============================================================================

func TestPullModelNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ollama = require("ai.ollama")
		local err = ollama.pull_model(nil, "llama2")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// show_model tests
// =============================================================================

func TestShowModelNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ollama = require("ai.ollama")
		local info, err = ollama.show_model(nil, "llama2")
		assert(info == nil, "info should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// copy_model tests
// =============================================================================

func TestCopyModelNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ollama = require("ai.ollama")
		local err = ollama.copy_model(nil, "source", "dest")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// delete_model tests
// =============================================================================

func TestDeleteModelNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local ollama = require("ai.ollama")
		local err = ollama.delete_model(nil, "model")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
