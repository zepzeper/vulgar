package slack

import (
	lua "github.com/yuin/gopher-lua"
)

// Usage: local err = slack.send_webhook(webhook_url, {text = "Hello!", channel = "#general"})
func luaSendWebhook(L *lua.LState) int {
	webhookURL := L.CheckString(1)
	opts := L.OptTable(2, nil)

	if webhookURL == "" {
		L.Push(lua.LString("webhook URL is required"))
		return 1
	}

	payload := &webhookPayload{}

	if opts != nil {
		if v := L.GetField(opts, "text"); v != lua.LNil {
			payload.Text = lua.LVAsString(v)
		}
		if v := L.GetField(opts, "channel"); v != lua.LNil {
			payload.Channel = lua.LVAsString(v)
		}
		if v := L.GetField(opts, "username"); v != lua.LNil {
			payload.Username = lua.LVAsString(v)
		}
		if v := L.GetField(opts, "icon_emoji"); v != lua.LNil {
			payload.IconEmoji = lua.LVAsString(v)
		}
		if v := L.GetField(opts, "icon_url"); v != lua.LNil {
			payload.IconURL = lua.LVAsString(v)
		}
		if v := L.GetField(opts, "blocks"); v != lua.LNil {
			if tbl, ok := v.(*lua.LTable); ok {
				payload.Blocks = luaTableToBlocks(tbl)
			}
		}
		if v := L.GetField(opts, "attachments"); v != lua.LNil {
			if tbl, ok := v.(*lua.LTable); ok {
				payload.Attachments = luaTableToBlocks(tbl)
			}
		}
	} else {
		L.Push(lua.LString("payload table required"))
		return 1
	}

	if err := sendWebhook(webhookURL, payload); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}
