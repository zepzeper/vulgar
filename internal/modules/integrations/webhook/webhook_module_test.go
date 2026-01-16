package webhook

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

// =============================================================================
// send tests
// =============================================================================

func TestSendBasic(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local webhook = require("integrations.webhook")
		
		-- Using httpbin for testing
		local response, err = webhook.send("https://httpbin.org/post", "test payload", {
			method = "POST",
			headers = {["Content-Type"] = "text/plain"}
		})
		
		assert(err == nil, "send should not error: " .. tostring(err))
		assert(response ~= nil, "response should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSendInvalidURL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local webhook = require("integrations.webhook")
		
		local response, err = webhook.send("not-a-valid-url", "payload")
		assert(response == nil, "response should be nil for invalid URL")
		assert(err ~= nil, "should error for invalid URL")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSendWithHeaders(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local webhook = require("integrations.webhook")
		
		local response, err = webhook.send("https://httpbin.org/post", "test", {
			method = "POST",
			headers = {
				["X-Custom-Header"] = "custom-value",
				["Authorization"] = "Bearer token123"
			}
		})
		
		assert(err == nil, "send with headers should not error: " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// send_json tests
// =============================================================================

func TestSendJSONBasic(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local webhook = require("integrations.webhook")
		
		local response, err = webhook.send_json("https://httpbin.org/post", {
			name = "Test",
			value = 123
		})
		
		assert(err == nil, "send_json should not error: " .. tostring(err))
		assert(response ~= nil, "response should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSendJSONInvalidURL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local webhook = require("integrations.webhook")
		
		local response, err = webhook.send_json("invalid-url", {test = true})
		assert(response == nil, "response should be nil for invalid URL")
		assert(err ~= nil, "should error for invalid URL")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSendJSONEmpty(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local webhook = require("integrations.webhook")
		
		local response, err = webhook.send_json("https://httpbin.org/post", {})
		assert(err == nil, "send_json with empty object should not error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// listen tests
// =============================================================================

func TestListenAndStop(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local webhook = require("integrations.webhook")
		
		local server, err = webhook.listen(0, function(req)
			return {status = 200, body = "ok"}
		end)
		
		assert(err == nil, "listen should not error: " .. tostring(err))
		assert(server ~= nil, "server should not be nil")
		
		-- Stop the server
		err = webhook.stop(server)
		assert(err == nil, "stop should not error")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestListenInvalidPort(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local webhook = require("integrations.webhook")
		
		local server, err = webhook.listen(-1, function(req)
			return {status = 200}
		end)
		-- Negative port should error
		assert(server == nil or err ~= nil, "should handle invalid port")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// verify tests
// =============================================================================

func TestVerifyValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local webhook = require("integrations.webhook")
		
		local payload = "test payload"
		local secret = "my-secret"
		
		-- First sign the payload
		local signature, err = webhook.sign(payload, secret, {algorithm = "sha256"})
		assert(err == nil, "sign should not error")
		
		-- Then verify it
		local valid, err = webhook.verify(payload, signature, secret, {algorithm = "sha256"})
		assert(err == nil, "verify should not error: " .. tostring(err))
		assert(valid == true, "signature should be valid")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestVerifyInvalid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local webhook = require("integrations.webhook")
		
		local valid, err = webhook.verify("payload", "invalid-signature", "secret")
		assert(valid == false, "should return false for invalid signature")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// sign tests
// =============================================================================

func TestSignPayload(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local webhook = require("integrations.webhook")
		
		local signature, err = webhook.sign("test payload", "secret", {algorithm = "sha256"})
		assert(err == nil, "sign should not error: " .. tostring(err))
		assert(signature ~= nil, "signature should not be nil")
		assert(#signature > 0, "signature should not be empty")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSignDifferentAlgorithms(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local webhook = require("integrations.webhook")
		
		local sig256, _ = webhook.sign("payload", "secret", {algorithm = "sha256"})
		local sig512, _ = webhook.sign("payload", "secret", {algorithm = "sha512"})
		
		-- Different algorithms should produce different signatures
		assert(sig256 ~= sig512, "different algorithms should produce different signatures")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// stop tests
// =============================================================================

func TestStopNil(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local webhook = require("integrations.webhook")
		local err = webhook.stop(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
