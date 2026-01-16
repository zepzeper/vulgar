package slack

import (
	lua "github.com/yuin/gopher-lua"
)

// Usage: local err = slack.react(client, "#channel", "1234567890.123456", "thumbsup")
func luaReact(L *lua.LState) int {
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
	emoji := L.CheckString(4)

	req := map[string]string{
		"channel":   channel,
		"timestamp": timestamp,
		"name":      emoji,
	}

	_, err := client.apiRequest("POST", "/reactions.add", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Client method wrapper
func luaClientReact(L *lua.LState) int {
	client := checkSlackClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	channel := L.CheckString(2)
	timestamp := L.CheckString(3)
	emoji := L.CheckString(4)

	req := map[string]string{
		"channel":   channel,
		"timestamp": timestamp,
		"name":      emoji,
	}

	_, err := client.apiRequest("POST", "/reactions.add", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}
