package discord

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
		local discord = require("integrations.discord")
		local client, err = discord.client({token = "test-token"})
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
		local discord = require("integrations.discord")
		local client, err = discord.client({})
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
		local discord = require("integrations.discord")
		local err = discord.send(nil, "channel_id", "Hello!")
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
		local discord = require("integrations.discord")
		local err = discord.send_webhook("not-a-url", {content = "Hello!"})
		assert(err ~= nil, "should error for invalid URL")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// send_embed tests
// =============================================================================

func TestSendEmbedNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local discord = require("integrations.discord")
		local err = discord.send_embed(nil, "channel_id", {
			title = "Test Embed",
			description = "Hello!"
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
		local discord = require("integrations.discord")
		local err = discord.upload_file(nil, "channel_id", {
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
		local discord = require("integrations.discord")
		local channels, err = discord.list_channels(nil, "guild_id")
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
		local discord = require("integrations.discord")
		local user, err = discord.get_user(nil, "user_id")
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
		local discord = require("integrations.discord")
		local err = discord.react(nil, "channel_id", "message_id", "thumbsup")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
