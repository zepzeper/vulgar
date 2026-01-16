package slack

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// Usage: local channels, err = slack.list_channels(client)
func luaListChannels(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return util.PushError(L, "client is required")
	}

	client := checkSlackClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	result, err := client.apiRequest("GET", "/conversations.list?types=public_channel,private_channel", nil)
	if err != nil {
		return util.PushError(L, "list channels failed: %v", err)
	}

	channels := L.NewTable()
	if chans, ok := result["channels"].([]interface{}); ok {
		for i, ch := range chans {
			if chMap, ok := ch.(map[string]interface{}); ok {
				chTbl := L.NewTable()
				if id, ok := chMap["id"].(string); ok {
					chTbl.RawSetString("id", lua.LString(id))
				}
				if name, ok := chMap["name"].(string); ok {
					chTbl.RawSetString("name", lua.LString(name))
				}
				if isPrivate, ok := chMap["is_private"].(bool); ok {
					chTbl.RawSetString("is_private", lua.LBool(isPrivate))
				}
				channels.RawSetInt(i+1, chTbl)
			}
		}
	}

	return util.PushSuccess(L, channels)
}

// Usage: local channel, err = slack.get_channel(client, "#channel")
// Or: local channel, err = slack.get_channel(client, "C1234567890")
func luaGetChannel(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return util.PushError(L, "client is required")
	}

	client := checkSlackClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	channelID := L.CheckString(2)

	result, err := client.apiRequest("GET", "/conversations.info?channel="+channelID, nil)
	if err != nil {
		return util.PushError(L, "get channel failed: %v", err)
	}

	channelTbl := L.NewTable()
	if channel, ok := result["channel"].(map[string]interface{}); ok {
		if id, ok := channel["id"].(string); ok {
			channelTbl.RawSetString("id", lua.LString(id))
		}
		if name, ok := channel["name"].(string); ok {
			channelTbl.RawSetString("name", lua.LString(name))
		}
		if isPrivate, ok := channel["is_private"].(bool); ok {
			channelTbl.RawSetString("is_private", lua.LBool(isPrivate))
		}
		if isArchived, ok := channel["is_archived"].(bool); ok {
			channelTbl.RawSetString("is_archived", lua.LBool(isArchived))
		}
		if topic, ok := channel["topic"].(map[string]interface{}); ok {
			if value, ok := topic["value"].(string); ok {
				channelTbl.RawSetString("topic", lua.LString(value))
			}
		}
		if purpose, ok := channel["purpose"].(map[string]interface{}); ok {
			if value, ok := purpose["value"].(string); ok {
				channelTbl.RawSetString("purpose", lua.LString(value))
			}
		}
		if numMembers, ok := channel["num_members"].(float64); ok {
			channelTbl.RawSetString("num_members", lua.LNumber(numMembers))
		}
	}

	return util.PushSuccess(L, channelTbl)
}

// Client method wrappers
func luaClientListChannels(L *lua.LState) int {
	client := checkSlackClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	result, err := client.apiRequest("GET", "/conversations.list?types=public_channel,private_channel", nil)
	if err != nil {
		return util.PushError(L, "list channels failed: %v", err)
	}

	channels := L.NewTable()
	if chans, ok := result["channels"].([]interface{}); ok {
		for i, ch := range chans {
			if chMap, ok := ch.(map[string]interface{}); ok {
				chTbl := L.NewTable()
				if id, ok := chMap["id"].(string); ok {
					chTbl.RawSetString("id", lua.LString(id))
				}
				if name, ok := chMap["name"].(string); ok {
					chTbl.RawSetString("name", lua.LString(name))
				}
				if isPrivate, ok := chMap["is_private"].(bool); ok {
					chTbl.RawSetString("is_private", lua.LBool(isPrivate))
				}
				channels.RawSetInt(i+1, chTbl)
			}
		}
	}

	return util.PushSuccess(L, channels)
}

func luaClientGetChannel(L *lua.LState) int {
	client := checkSlackClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	channelID := L.CheckString(2)

	result, err := client.apiRequest("GET", "/conversations.info?channel="+channelID, nil)
	if err != nil {
		return util.PushError(L, "get channel failed: %v", err)
	}

	channelTbl := L.NewTable()
	if channel, ok := result["channel"].(map[string]interface{}); ok {
		if id, ok := channel["id"].(string); ok {
			channelTbl.RawSetString("id", lua.LString(id))
		}
		if name, ok := channel["name"].(string); ok {
			channelTbl.RawSetString("name", lua.LString(name))
		}
		if isPrivate, ok := channel["is_private"].(bool); ok {
			channelTbl.RawSetString("is_private", lua.LBool(isPrivate))
		}
		if isArchived, ok := channel["is_archived"].(bool); ok {
			channelTbl.RawSetString("is_archived", lua.LBool(isArchived))
		}
		if topic, ok := channel["topic"].(map[string]interface{}); ok {
			if value, ok := topic["value"].(string); ok {
				channelTbl.RawSetString("topic", lua.LString(value))
			}
		}
		if purpose, ok := channel["purpose"].(map[string]interface{}); ok {
			if value, ok := purpose["value"].(string); ok {
				channelTbl.RawSetString("purpose", lua.LString(value))
			}
		}
		if numMembers, ok := channel["num_members"].(float64); ok {
			channelTbl.RawSetString("num_members", lua.LNumber(numMembers))
		}
	}

	return util.PushSuccess(L, channelTbl)
}
