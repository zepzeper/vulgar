package slack

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// Usage: local err = slack.send(client, "#channel", "Hello!")
// Or: local err = slack.send(client, "#channel", "Hello!", {thread_ts = "1234567890.123456"})
func luaSend(L *lua.LState) int {
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
	text := L.CheckString(3)
	opts := L.OptTable(4, nil)

	req := chatPostMessageRequest{
		Channel: channel,
		Text:    text,
		Mrkdwn:  true,
	}

	// Support thread replies
	if opts != nil {
		if v := L.GetField(opts, "thread_ts"); v != lua.LNil {
			req.ThreadTS = lua.LVAsString(v)
		}
	}

	_, err := client.apiRequest("POST", "/chat.postMessage", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: local err = slack.send_blocks(client, "#channel", blocks)
// Or: local err = slack.send_blocks(client, "#channel", blocks, {thread_ts = "1234567890.123456"})
func luaSendBlocks(L *lua.LState) int {
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
	blocksTbl := L.CheckTable(3)
	opts := L.OptTable(4, nil)

	blocks := luaTableToBlocks(blocksTbl)

	req := chatPostMessageRequest{
		Channel: channel,
		Blocks:  blocks,
	}

	// Support thread replies
	if opts != nil {
		if v := L.GetField(opts, "thread_ts"); v != lua.LNil {
			req.ThreadTS = lua.LVAsString(v)
		}
	}

	_, err := client.apiRequest("POST", "/chat.postMessage", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: local err = slack.upload_file(client, "#channel", {filename = "test.txt", content = "Hello"})
func luaUploadFile(L *lua.LState) int {
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
	opts := L.OptTable(3, nil)
	if opts == nil {
		L.Push(lua.LString("file options required"))
		return 1
	}

	filename := ""
	content := ""
	title := ""
	comment := ""

	if v := L.GetField(opts, "filename"); v != lua.LNil {
		filename = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "content"); v != lua.LNil {
		content = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "title"); v != lua.LNil {
		title = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "comment"); v != lua.LNil {
		comment = lua.LVAsString(v)
	}

	if filename == "" {
		filename = "file.txt"
	}

	// Use httpclient's multipart builder
	mb := client.client.NewMultipartRequest("POST", "/files.upload").
		Field("channels", channel).
		Field("filename", filename).
		FileFromBytes("file", filename, []byte(content))

	if title != "" {
		mb.Field("title", title)
	}
	if comment != "" {
		mb.Field("initial_comment", comment)
	}

	resp, err := mb.Do()
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("upload failed: %v", err)))
		return 1
	}

	_, err = resp.CheckSlack()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: local err = slack.update_message(client, "#channel", "1234567890.123456", "Updated text")
// Or: local err = slack.update_message(client, "#channel", "1234567890.123456", nil, {blocks = {...}})
func luaUpdateMessage(L *lua.LState) int {
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

	req := map[string]interface{}{
		"channel": channel,
		"ts":      timestamp,
	}

	// Text or blocks
	if L.Get(4) != lua.LNil {
		if text := L.CheckString(4); text != "" {
			req["text"] = text
		}
	}

	opts := L.OptTable(5, nil)
	if opts != nil {
		if v := L.GetField(opts, "text"); v != lua.LNil {
			req["text"] = lua.LVAsString(v)
		}
		if v := L.GetField(opts, "blocks"); v != lua.LNil {
			if tbl, ok := v.(*lua.LTable); ok {
				req["blocks"] = luaTableToBlocks(tbl)
			}
		}
	}

	_, err := client.apiRequest("POST", "/chat.update", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: local err = slack.delete_message(client, "#channel", "1234567890.123456")
func luaDeleteMessage(L *lua.LState) int {
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
		"channel": channel,
		"ts":      timestamp,
	}

	_, err := client.apiRequest("POST", "/chat.delete", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Client method wrappers
func luaClientSend(L *lua.LState) int {
	client := checkSlackClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	channel := L.CheckString(2)
	text := L.CheckString(3)
	opts := L.OptTable(4, nil)

	req := chatPostMessageRequest{
		Channel: channel,
		Text:    text,
		Mrkdwn:  true,
	}

	// Support thread replies
	if opts != nil {
		if v := L.GetField(opts, "thread_ts"); v != lua.LNil {
			req.ThreadTS = lua.LVAsString(v)
		}
	}

	_, err := client.apiRequest("POST", "/chat.postMessage", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

func luaClientSendBlocks(L *lua.LState) int {
	client := checkSlackClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	channel := L.CheckString(2)
	blocksTbl := L.CheckTable(3)
	opts := L.OptTable(4, nil)

	blocks := luaTableToBlocks(blocksTbl)

	req := chatPostMessageRequest{
		Channel: channel,
		Blocks:  blocks,
	}

	// Support thread replies
	if opts != nil {
		if v := L.GetField(opts, "thread_ts"); v != lua.LNil {
			req.ThreadTS = lua.LVAsString(v)
		}
	}

	_, err := client.apiRequest("POST", "/chat.postMessage", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

func luaClientUploadFile(L *lua.LState) int {
	client := checkSlackClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	channel := L.CheckString(2)
	opts := L.OptTable(3, nil)
	if opts == nil {
		L.Push(lua.LString("file options required"))
		return 1
	}

	filename := ""
	content := ""
	title := ""
	comment := ""

	if v := L.GetField(opts, "filename"); v != lua.LNil {
		filename = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "content"); v != lua.LNil {
		content = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "title"); v != lua.LNil {
		title = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "comment"); v != lua.LNil {
		comment = lua.LVAsString(v)
	}

	if filename == "" {
		filename = "file.txt"
	}

	mb := client.client.NewMultipartRequest("POST", "/files.upload").
		Field("channels", channel).
		Field("filename", filename).
		FileFromBytes("file", filename, []byte(content))

	if title != "" {
		mb.Field("title", title)
	}
	if comment != "" {
		mb.Field("initial_comment", comment)
	}

	resp, err := mb.Do()
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("upload failed: %v", err)))
		return 1
	}

	_, err = resp.CheckSlack()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

func luaClientUpdateMessage(L *lua.LState) int {
	client := checkSlackClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	channel := L.CheckString(2)
	timestamp := L.CheckString(3)

	req := map[string]interface{}{
		"channel": channel,
		"ts":      timestamp,
	}

	// Text or blocks
	if L.Get(4) != lua.LNil {
		if text := L.CheckString(4); text != "" {
			req["text"] = text
		}
	}

	opts := L.OptTable(5, nil)
	if opts != nil {
		if v := L.GetField(opts, "text"); v != lua.LNil {
			req["text"] = lua.LVAsString(v)
		}
		if v := L.GetField(opts, "blocks"); v != lua.LNil {
			if tbl, ok := v.(*lua.LTable); ok {
				req["blocks"] = luaTableToBlocks(tbl)
			}
		}
	}

	_, err := client.apiRequest("POST", "/chat.update", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

func luaClientDeleteMessage(L *lua.LState) int {
	client := checkSlackClient(L, 1)
	if client == nil {
		L.Push(lua.LString("invalid client"))
		return 1
	}

	channel := L.CheckString(2)
	timestamp := L.CheckString(3)

	req := map[string]string{
		"channel": channel,
		"ts":      timestamp,
	}

	_, err := client.apiRequest("POST", "/chat.delete", req)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}
