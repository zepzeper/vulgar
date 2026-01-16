package slack

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
		local slack = require("integrations.slack")
		local client, err = slack.client({token = "xoxb-xxx"})
		assert(err == nil, "client should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestClientMissingToken(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local slack = require("integrations.slack")
		local client, err = slack.client({})
		assert(client == nil or err ~= nil, "should error with missing token")
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
		local slack = require("integrations.slack")
		local err = slack.send(nil, "#channel", "Hello!")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// send_webhook tests
// =============================================================================

func TestSendWebhookInvalidURL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local slack = require("integrations.slack")
		local err = slack.send_webhook("not-a-url", {text = "Hello!"})
		assert(err ~= nil, "should error for invalid URL")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// send_blocks tests
// =============================================================================

func TestSendBlocksNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local slack = require("integrations.slack")
		local err = slack.send_blocks(nil, "#channel", {
			{type = "section", text = {type = "mrkdwn", text = "Hello!"}}
		})
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// upload_file tests
// =============================================================================

func TestUploadFileNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local slack = require("integrations.slack")
		local err = slack.upload_file(nil, "#channel", {
			filename = "test.txt",
			content = "Hello World"
		})
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list_channels tests
// =============================================================================

func TestListChannelsNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local slack = require("integrations.slack")
		local channels, err = slack.list_channels(nil)
		assert(channels == nil, "channels should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// get_user tests
// =============================================================================

func TestGetUserNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local slack = require("integrations.slack")
		local user, err = slack.get_user(nil, "U12345")
		assert(user == nil, "user should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// react tests
// =============================================================================

func TestReactNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local slack = require("integrations.slack")
		local err = slack.react(nil, "#channel", "1234567890.123456", "thumbsup")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
