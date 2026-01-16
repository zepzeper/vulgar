package twilio

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
// client tests
// =============================================================================

func TestClientCreate(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local twilio = require("integrations.twilio")
		local client, err = twilio.client({
			account_sid = "ACxxx",
			auth_token = "xxx"
		})
		assert(err == nil, "client should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestClientMissingCredentials(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local twilio = require("integrations.twilio")
		local client, err = twilio.client({})
		assert(client == nil or err ~= nil, "should error with missing credentials")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// send_sms tests
// =============================================================================

func TestSendSMSNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local twilio = require("integrations.twilio")
		local msg, err = twilio.send_sms(nil, {
			from = "+15551234567",
			to = "+15557654321",
			body = "Test message"
		})
		assert(msg == nil, "msg should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSendSMSMissingParams(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local twilio = require("integrations.twilio")
		local client, _ = twilio.client({
			account_sid = "ACxxx",
			auth_token = "xxx"
		})
		
		local msg, err = twilio.send_sms(client, {})
		assert(msg == nil or err ~= nil, "should error with missing params")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// send_mms tests
// =============================================================================

func TestSendMMSNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local twilio = require("integrations.twilio")
		local msg, err = twilio.send_mms(nil, {
			from = "+15551234567",
			to = "+15557654321",
			body = "Test",
			media_url = "https://example.com/image.png"
		})
		assert(msg == nil, "msg should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// make_call tests
// =============================================================================

func TestMakeCallNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local twilio = require("integrations.twilio")
		local call, err = twilio.make_call(nil, {
			from = "+15551234567",
			to = "+15557654321",
			url = "http://example.com/twiml"
		})
		assert(call == nil, "call should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// get_message tests
// =============================================================================

func TestGetMessageNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local twilio = require("integrations.twilio")
		local msg, err = twilio.get_message(nil, "SMxxx")
		assert(msg == nil, "msg should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list_messages tests
// =============================================================================

func TestListMessagesNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local twilio = require("integrations.twilio")
		local messages, err = twilio.list_messages(nil)
		assert(messages == nil, "messages should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// lookup tests
// =============================================================================

func TestLookupNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local twilio = require("integrations.twilio")
		local info, err = twilio.lookup(nil, "+15551234567")
		assert(info == nil, "info should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// verify_signature tests
// =============================================================================

func TestVerifySignature(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local twilio = require("integrations.twilio")
		local valid, err = twilio.verify_signature("url", "params", "signature", "auth_token")
		-- Just testing that function exists and is callable
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
