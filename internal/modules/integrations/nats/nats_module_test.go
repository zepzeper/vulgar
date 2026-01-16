package nats

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

func skipIfNoNATS(t *testing.T) {
	if os.Getenv("NATS_TEST_URL") == "" {
		t.Skip("NATS_TEST_URL not set, skipping integration test")
	}
}

// =============================================================================
// connect tests
// =============================================================================

func TestConnectMissingURL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local nats = require("integrations.nats")
		local conn, err = nats.connect({})
		assert(conn == nil or err ~= nil, "should error with missing URL")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectInvalidURL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local nats = require("integrations.nats")
		local conn, err = nats.connect({url = "nats://invalid-host:4222"})
		assert(conn == nil, "conn should be nil for invalid host")
		assert(err ~= nil, "should error for invalid host")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectValid(t *testing.T) {
	skipIfNoNATS(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("url", lua.LString(os.Getenv("NATS_TEST_URL")))

	err := L.DoString(`
		local nats = require("integrations.nats")
		local conn, err = nats.connect({url = url})
		assert(err == nil, "connect should not error: " .. tostring(err))
		assert(conn ~= nil, "conn should not be nil")
		nats.close(conn)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// publish tests
// =============================================================================

func TestPublishNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local nats = require("integrations.nats")
		local err = nats.publish(nil, "subject", "message")
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestPublish(t *testing.T) {
	skipIfNoNATS(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("url", lua.LString(os.Getenv("NATS_TEST_URL")))

	err := L.DoString(`
		local nats = require("integrations.nats")
		local conn, _ = nats.connect({url = url})
		
		local err = nats.publish(conn, "test.subject", "test message")
		assert(err == nil, "publish should not error: " .. tostring(err))
		
		nats.close(conn)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// subscribe tests
// =============================================================================

func TestSubscribeNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local nats = require("integrations.nats")
		local sub, err = nats.subscribe(nil, "subject", function() end)
		assert(sub == nil, "sub should be nil")
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSubscribe(t *testing.T) {
	skipIfNoNATS(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("url", lua.LString(os.Getenv("NATS_TEST_URL")))

	err := L.DoString(`
		local nats = require("integrations.nats")
		local conn, _ = nats.connect({url = url})
		
		local sub, err = nats.subscribe(conn, "test.subject", function(msg)
			-- Handle message
		end)
		assert(err == nil, "subscribe should not error: " .. tostring(err))
		assert(sub ~= nil, "sub should not be nil")
		
		nats.unsubscribe(sub)
		nats.close(conn)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// request tests
// =============================================================================

func TestRequestNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local nats = require("integrations.nats")
		local response, err = nats.request(nil, "subject", "message", {timeout = 1})
		assert(response == nil, "response should be nil")
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// queue_subscribe tests
// =============================================================================

func TestQueueSubscribeNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local nats = require("integrations.nats")
		local sub, err = nats.queue_subscribe(nil, "subject", "queue", function() end)
		assert(sub == nil, "sub should be nil")
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// unsubscribe tests
// =============================================================================

func TestUnsubscribeNil(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local nats = require("integrations.nats")
		local err = nats.unsubscribe(nil)
		-- Should handle nil gracefully
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
		local nats = require("integrations.nats")
		local err = nats.close(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
