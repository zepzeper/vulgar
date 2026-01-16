package kafka

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

func skipIfNoKafka(t *testing.T) {
	if os.Getenv("KAFKA_TEST_BROKERS") == "" {
		t.Skip("KAFKA_TEST_BROKERS not set, skipping integration test")
	}
}

// =============================================================================
// connect tests
// =============================================================================

func TestConnectMissingBrokers(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local kafka = require("integrations.kafka")
		local client, err = kafka.connect({})
		assert(client == nil or err ~= nil, "should error with missing brokers")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectInvalidBrokers(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local kafka = require("integrations.kafka")
		local client, err = kafka.connect({brokers = {"invalid-host:9092"}})
		assert(client == nil, "client should be nil for invalid brokers")
		assert(err ~= nil, "should error for invalid brokers")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectValid(t *testing.T) {
	skipIfNoKafka(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("brokers", lua.LString(os.Getenv("KAFKA_TEST_BROKERS")))

	err := L.DoString(`
		local kafka = require("integrations.kafka")
		local client, err = kafka.connect({brokers = {brokers}})
		assert(err == nil, "connect should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
		kafka.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// produce tests
// =============================================================================

func TestProduceNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local kafka = require("integrations.kafka")
		local err = kafka.produce(nil, "topic", "message")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestProduce(t *testing.T) {
	skipIfNoKafka(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("brokers", lua.LString(os.Getenv("KAFKA_TEST_BROKERS")))
	L.SetGlobal("topic", lua.LString(os.Getenv("KAFKA_TEST_TOPIC")))

	err := L.DoString(`
		local kafka = require("integrations.kafka")
		local client, _ = kafka.connect({brokers = {brokers}})
		
		local err = kafka.produce(client, topic, "test message")
		assert(err == nil, "produce should not error: " .. tostring(err))
		
		kafka.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestProduceWithKey(t *testing.T) {
	skipIfNoKafka(t)
	L := newTestState()
	defer L.Close()

	L.SetGlobal("brokers", lua.LString(os.Getenv("KAFKA_TEST_BROKERS")))
	L.SetGlobal("topic", lua.LString(os.Getenv("KAFKA_TEST_TOPIC")))

	err := L.DoString(`
		local kafka = require("integrations.kafka")
		local client, _ = kafka.connect({brokers = {brokers}})
		
		local err = kafka.produce(client, topic, "test message", {key = "my-key"})
		assert(err == nil, "produce with key should not error: " .. tostring(err))
		
		kafka.close(client)
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// consume tests
// =============================================================================

func TestConsumeNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local kafka = require("integrations.kafka")
		local messages, err = kafka.consume(nil, "topic", {timeout = 1})
		assert(messages == nil, "messages should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// subscribe tests
// =============================================================================

func TestSubscribeNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local kafka = require("integrations.kafka")
		local err = kafka.subscribe(nil, {"topic"}, function() end)
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
		local kafka = require("integrations.kafka")
		local err = kafka.close(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
