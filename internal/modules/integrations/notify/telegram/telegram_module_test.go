package telegram

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
		local telegram = require("integrations.telegram")
		local client, err = telegram.client({token = "123456:ABC-xxx"})
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
		local telegram = require("integrations.telegram")
		local client, err = telegram.client({})
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
		local telegram = require("integrations.telegram")
		local err = telegram.send(nil, "chat_id", "Hello!")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// send_photo tests
// =============================================================================

func TestSendPhotoNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local telegram = require("integrations.telegram")
		local err = telegram.send_photo(nil, "chat_id", "https://example.com/photo.jpg")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// send_document tests
// =============================================================================

func TestSendDocumentNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local telegram = require("integrations.telegram")
		local err = telegram.send_document(nil, "chat_id", {
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
// send_sticker tests
// =============================================================================

func TestSendStickerNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local telegram = require("integrations.telegram")
		local err = telegram.send_sticker(nil, "chat_id", "sticker_id")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// get_updates tests
// =============================================================================

func TestGetUpdatesNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local telegram = require("integrations.telegram")
		local updates, err = telegram.get_updates(nil)
		assert(updates == nil, "updates should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// set_webhook tests
// =============================================================================

func TestSetWebhookNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local telegram = require("integrations.telegram")
		local err = telegram.set_webhook(nil, "https://example.com/webhook")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// get_me tests
// =============================================================================

func TestGetMeNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local telegram = require("integrations.telegram")
		local me, err = telegram.get_me(nil)
		assert(me == nil, "me should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// reply tests
// =============================================================================

func TestReplyNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local telegram = require("integrations.telegram")
		local err = telegram.reply(nil, "chat_id", "message_id", "Reply text")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// edit_message tests
// =============================================================================

func TestEditMessageNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local telegram = require("integrations.telegram")
		local err = telegram.edit_message(nil, "chat_id", "message_id", "New text")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
