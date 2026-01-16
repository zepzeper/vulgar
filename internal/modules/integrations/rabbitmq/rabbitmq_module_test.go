package rabbitmq

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

func skipIfNoRabbitMQ(t *testing.T) {
	if os.Getenv("RABBITMQ_TEST_URI") == "" {
		t.Skip("RABBITMQ_TEST_URI not set, skipping integration test")
	}
}

// =============================================================================
// connect tests
// =============================================================================

func TestConnectMissingURI(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local rabbitmq = require("integrations.rabbitmq")
		local conn, err = rabbitmq.connect({})
		assert(conn == nil or err ~= nil, "should error with missing URI")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectInvalidURI(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local rabbitmq = require("integrations.rabbitmq")
		local conn, err = rabbitmq.connect({uri = "amqp://invalid-host:5672"})
		assert(conn == nil, "conn should be nil for invalid host")
		assert(err ~= nil, "should error for invalid host")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectValid(t *testing.T) {
	skipIfNoRabbitMQ(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("uri", lua.LString(os.Getenv("RABBITMQ_TEST_URI")))

	err := L.DoString(`
		local rabbitmq = require("integrations.rabbitmq")
		local conn, err = rabbitmq.connect({uri = uri})
		assert(err == nil, "connect should not error: " .. tostring(err))
		assert(conn ~= nil, "conn should not be nil")
		rabbitmq.close(conn)
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
		local rabbitmq = require("integrations.rabbitmq")
		local err = rabbitmq.publish(nil, "exchange", "routing_key", "message")
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestPublish(t *testing.T) {
	skipIfNoRabbitMQ(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("uri", lua.LString(os.Getenv("RABBITMQ_TEST_URI")))

	err := L.DoString(`
		local rabbitmq = require("integrations.rabbitmq")
		local conn, _ = rabbitmq.connect({uri = uri})
		
		local err = rabbitmq.publish(conn, "", "test_queue", "test message")
		assert(err == nil, "publish should not error: " .. tostring(err))
		
		rabbitmq.close(conn)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// consume tests
// =============================================================================

func TestConsumeNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local rabbitmq = require("integrations.rabbitmq")
		local err = rabbitmq.consume(nil, "queue", function() end)
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// declare_queue tests
// =============================================================================

func TestDeclareQueueNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local rabbitmq = require("integrations.rabbitmq")
		local err = rabbitmq.declare_queue(nil, "queue")
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestDeclareQueue(t *testing.T) {
	skipIfNoRabbitMQ(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("uri", lua.LString(os.Getenv("RABBITMQ_TEST_URI")))

	err := L.DoString(`
		local rabbitmq = require("integrations.rabbitmq")
		local conn, _ = rabbitmq.connect({uri = uri})
		
		local err = rabbitmq.declare_queue(conn, "test_queue", {durable = true})
		assert(err == nil, "declare_queue should not error: " .. tostring(err))
		
		rabbitmq.close(conn)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// declare_exchange tests
// =============================================================================

func TestDeclareExchangeNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local rabbitmq = require("integrations.rabbitmq")
		local err = rabbitmq.declare_exchange(nil, "exchange", "direct")
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// bind_queue tests
// =============================================================================

func TestBindQueueNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local rabbitmq = require("integrations.rabbitmq")
		local err = rabbitmq.bind_queue(nil, "queue", "exchange", "routing_key")
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// ack / nack tests
// =============================================================================

func TestAckNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local rabbitmq = require("integrations.rabbitmq")
		local err = rabbitmq.ack(nil, 1)
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestNackNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local rabbitmq = require("integrations.rabbitmq")
		local err = rabbitmq.nack(nil, 1)
		assert(err ~= nil, "should error without connection")
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
		local rabbitmq = require("integrations.rabbitmq")
		local err = rabbitmq.close(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
