package slack

import (
	lua "github.com/yuin/gopher-lua"
)

// Usage: local err = slack.pin_message(client, "#channel", "1234567890.123456")
func luaPinMessage(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LString("client is required"))
		return 1
	}

	client := checkSlackClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	channel := L.CheckString(2)
	timestamp := L.CheckString(3)

	req := map[string]string{
		"channel":   channel,
		"timestamp": timestamp,
	}

	_, err := client.apiRequest("POST", "/pins.add", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: local err = slack.unpin_message(client, "#channel", "1234567890.123456")
func luaUnpinMessage(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		L.Push(lua.LString("client is required"))
		return 1
	}

	client := checkSlackClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	channel := L.CheckString(2)
	timestamp := L.CheckString(3)

	req := map[string]string{
		"channel":   channel,
		"timestamp": timestamp,
	}

	_, err := client.apiRequest("POST", "/pins.remove", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Client method wrappers
func luaClientPinMessage(L *lua.LState) int {
	client := checkSlackClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	channel := L.CheckString(2)
	timestamp := L.CheckString(3)

	req := map[string]string{
		"channel":   channel,
		"timestamp": timestamp,
	}

	_, err := client.apiRequest("POST", "/pins.add", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

func luaClientUnpinMessage(L *lua.LState) int {
	client := checkSlackClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	channel := L.CheckString(2)
	timestamp := L.CheckString(3)

	req := map[string]string{
		"channel":   channel,
		"timestamp": timestamp,
	}

	_, err := client.apiRequest("POST", "/pins.remove", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}
