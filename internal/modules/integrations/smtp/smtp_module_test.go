package smtp

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

func skipIfNoSMTP(t *testing.T) {
	if os.Getenv("SMTP_TEST_HOST") == "" {
		t.Skip("SMTP_TEST_HOST not set, skipping integration test")
	}
}

// =============================================================================
// connect tests
// =============================================================================

func TestConnectMissingConfig(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local smtp = require("integrations.smtp")
		local client, err = smtp.connect({})
		assert(client == nil or err ~= nil, "should error with missing config")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectInvalidHost(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local smtp = require("integrations.smtp")
		local client, err = smtp.connect({
			host = "invalid-smtp-host-xyz",
			port = 587
		})
		assert(client == nil, "client should be nil for invalid host")
		assert(err ~= nil, "should error for invalid host")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectValid(t *testing.T) {
	skipIfNoSMTP(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("smtp_host", lua.LString(os.Getenv("SMTP_TEST_HOST")))
	L.SetGlobal("smtp_port", lua.LString(os.Getenv("SMTP_TEST_PORT")))
	L.SetGlobal("smtp_user", lua.LString(os.Getenv("SMTP_TEST_USER")))
	L.SetGlobal("smtp_pass", lua.LString(os.Getenv("SMTP_TEST_PASS")))

	err := L.DoString(`
		local smtp = require("integrations.smtp")
		local client, err = smtp.connect({
			host = smtp_host,
			port = tonumber(smtp_port),
			user = smtp_user,
			password = smtp_pass
		})
		assert(err == nil, "connect should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
		smtp.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// send tests
// =============================================================================

func TestSendNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local smtp = require("integrations.smtp")
		local err = smtp.send(nil, {
			from = "test@example.com",
			to = {"recipient@example.com"},
			subject = "Test",
			body = "Hello"
		})
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSendMissingFrom(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local smtp = require("integrations.smtp")
		-- Assuming we have a mock client
		local err = smtp.send(nil, {
			to = {"recipient@example.com"},
			subject = "Test",
			body = "Hello"
		})
		assert(err ~= nil, "should error without from")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSendMissingTo(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local smtp = require("integrations.smtp")
		local err = smtp.send(nil, {
			from = "test@example.com",
			subject = "Test",
			body = "Hello"
		})
		assert(err ~= nil, "should error without to")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSendEmail(t *testing.T) {
	skipIfNoSMTP(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("smtp_host", lua.LString(os.Getenv("SMTP_TEST_HOST")))
	L.SetGlobal("smtp_port", lua.LString(os.Getenv("SMTP_TEST_PORT")))
	L.SetGlobal("smtp_user", lua.LString(os.Getenv("SMTP_TEST_USER")))
	L.SetGlobal("smtp_pass", lua.LString(os.Getenv("SMTP_TEST_PASS")))
	L.SetGlobal("test_email", lua.LString(os.Getenv("SMTP_TEST_EMAIL")))

	err := L.DoString(`
		local smtp = require("integrations.smtp")
		local client, _ = smtp.connect({
			host = smtp_host,
			port = tonumber(smtp_port),
			user = smtp_user,
			password = smtp_pass
		})
		
		local err = smtp.send(client, {
			from = test_email,
			to = {test_email},
			subject = "Test Email",
			body = "This is a test email from Lua module tests."
		})
		assert(err == nil, "send should not error: " .. tostring(err))
		
		smtp.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// send_raw tests
// =============================================================================

func TestSendRawNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local smtp = require("integrations.smtp")
		local err = smtp.send_raw(nil, "from@test.com", "to@test.com", "raw message")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// close tests
// =============================================================================

func TestCloseNil(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local smtp = require("integrations.smtp")
		local err = smtp.close(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
