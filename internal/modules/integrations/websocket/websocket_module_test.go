package websocket

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
// connect tests
// =============================================================================

func TestConnectInvalidURL(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local websocket = require("integrations.websocket")
		local conn, err = websocket.connect("not-a-valid-url")
		assert(conn == nil, "conn should be nil for invalid URL")
		assert(err ~= nil, "should error for invalid URL")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectValid(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local websocket = require("integrations.websocket")
		-- Using a public echo server
		local conn, err = websocket.connect("wss://echo.websocket.org")
		
		if err == nil then
			assert(conn ~= nil, "conn should not be nil")
			websocket.close(conn)
		end
		-- May fail if echo server is down, which is OK
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestConnectWithHeaders(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local websocket = require("integrations.websocket")
		local conn, err = websocket.connect("wss://echo.websocket.org", {
			headers = {["Authorization"] = "Bearer token"}
		})
		
		if conn then
			websocket.close(conn)
		end
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// send tests
// =============================================================================

func TestSendNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local websocket = require("integrations.websocket")
		local err = websocket.send(nil, "message")
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestSendJSON(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local websocket = require("integrations.websocket")
		local conn, connErr = websocket.connect("wss://echo.websocket.org")
		
		if connErr == nil then
			local err = websocket.send_json(conn, {message = "hello"})
			assert(err == nil, "send_json should not error: " .. tostring(err))
			websocket.close(conn)
		end
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// receive tests
// =============================================================================

func TestReceiveNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local websocket = require("integrations.websocket")
		local msg, err = websocket.receive(nil)
		assert(msg == nil, "msg should be nil")
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// on_message tests
// =============================================================================

func TestOnMessageNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local websocket = require("integrations.websocket")
		local err = websocket.on_message(nil, function() end)
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// on_close tests
// =============================================================================

func TestOnCloseNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local websocket = require("integrations.websocket")
		local err = websocket.on_close(nil, function() end)
		assert(err ~= nil, "should error without connection")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// on_error tests
// =============================================================================

func TestOnErrorNoConnection(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local websocket = require("integrations.websocket")
		local err = websocket.on_error(nil, function() end)
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
		local websocket = require("integrations.websocket")
		local err = websocket.close(nil)
		-- Should handle nil gracefully
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
