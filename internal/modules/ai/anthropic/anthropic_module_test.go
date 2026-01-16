package anthropic

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
	if os.Getenv("ANTHROPIC_TEST_API_KEY") == "" {
		t.Skip("ANTHROPIC_TEST_API_KEY not set, skipping integration test")
	}
}

// =============================================================================
// client tests
// =============================================================================

func TestClientCreate(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local anthropic = require("ai.anthropic")
		local client, err = anthropic.client({api_key = "test-key"})
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
		local anthropic = require("ai.anthropic")
		local client, err = anthropic.client({})
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
		local anthropic = require("ai.anthropic")
		local response, err = anthropic.chat(nil, {
			model = "claude-3-sonnet-20240229",
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

	L.SetGlobal("api_key", lua.LString(os.Getenv("ANTHROPIC_TEST_API_KEY")))

	err := L.DoString(`
		local anthropic = require("ai.anthropic")
		local client, _ = anthropic.client({api_key = api_key})
		
		local response, err = anthropic.chat(client, {
			model = "claude-3-sonnet-20240229",
			max_tokens = 100,
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
		local anthropic = require("ai.anthropic")
		local response, err = anthropic.complete(nil, {
			model = "claude-3-sonnet-20240229",
			prompt = "Human: Hello\n\nAssistant:"
		})
		assert(response == nil, "response should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// stream tests
// =============================================================================

func TestStreamNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local anthropic = require("ai.anthropic")
		local stream, err = anthropic.stream(nil, {
			model = "claude-3-sonnet-20240229",
			messages = {{role = "user", content = "Hello"}}
		}, function() end)
		assert(stream == nil, "stream should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// count_tokens tests
// =============================================================================

func TestCountTokens(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local anthropic = require("ai.anthropic")
		local count, err = anthropic.count_tokens("Hello, how are you?")
		-- May or may not need a client
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// with_system tests
// =============================================================================

func TestWithSystem(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local anthropic = require("ai.anthropic")
		local messages = anthropic.with_system("You are helpful", {
			{role = "user", content = "Hello"}
		})
		-- Should prepend system message
		assert(messages ~= nil, "messages should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
