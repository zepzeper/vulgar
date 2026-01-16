package slack

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// Usage: local user, err = slack.get_user(client, "U12345")
func luaGetUser(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return util.PushError(L, "client is required")
	}

	client := checkSlackClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	userID := L.CheckString(2)

	result, err := client.apiRequest("GET", "/users.info?user="+userID, nil)
	if err != nil {
		return util.PushError(L, "get user failed: %v", err)
	}

	userTbl := L.NewTable()
	if user, ok := result["user"].(map[string]interface{}); ok {
		if id, ok := user["id"].(string); ok {
			userTbl.RawSetString("id", lua.LString(id))
		}
		if name, ok := user["name"].(string); ok {
			userTbl.RawSetString("name", lua.LString(name))
		}
		if realName, ok := user["real_name"].(string); ok {
			userTbl.RawSetString("real_name", lua.LString(realName))
		}
		if profile, ok := user["profile"].(map[string]interface{}); ok {
			if email, ok := profile["email"].(string); ok {
				userTbl.RawSetString("email", lua.LString(email))
			}
		}
	}

	return util.PushSuccess(L, userTbl)
}

// Usage: local users, err = slack.list_users(client)
// Or: local users, err = slack.list_users(client, {limit = 100})
func luaListUsers(L *lua.LState) int {
	if L.Get(1) == lua.LNil {
		return util.PushError(L, "client is required")
	}

	client := checkSlackClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	opts := L.OptTable(2, nil)
	limit := ""
	if opts != nil {
		if v := L.GetField(opts, "limit"); v != lua.LNil {
			limit = fmt.Sprintf("&limit=%d", int(lua.LVAsNumber(v)))
		}
	}

	endpoint := "/users.list" + limit
	result, err := client.apiRequest("GET", endpoint, nil)
	if err != nil {
		return util.PushError(L, "list users failed: %v", err)
	}

	users := L.NewTable()
	if usrs, ok := result["members"].([]interface{}); ok {
		for i, u := range usrs {
			if userMap, ok := u.(map[string]interface{}); ok {
				userTbl := L.NewTable()
				if id, ok := userMap["id"].(string); ok {
					userTbl.RawSetString("id", lua.LString(id))
				}
				if name, ok := userMap["name"].(string); ok {
					userTbl.RawSetString("name", lua.LString(name))
				}
				if realName, ok := userMap["real_name"].(string); ok {
					userTbl.RawSetString("real_name", lua.LString(realName))
				}
				if deleted, ok := userMap["deleted"].(bool); ok {
					userTbl.RawSetString("deleted", lua.LBool(deleted))
				}
				if isBot, ok := userMap["is_bot"].(bool); ok {
					userTbl.RawSetString("is_bot", lua.LBool(isBot))
				}
				if profile, ok := userMap["profile"].(map[string]interface{}); ok {
					if email, ok := profile["email"].(string); ok {
						userTbl.RawSetString("email", lua.LString(email))
					}
					if displayName, ok := profile["display_name"].(string); ok {
						userTbl.RawSetString("display_name", lua.LString(displayName))
					}
				}
				users.RawSetInt(i+1, userTbl)
			}
		}
	}

	return util.PushSuccess(L, users)
}

// Client method wrappers
func luaClientGetUser(L *lua.LState) int {
	client := checkSlackClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	userID := L.CheckString(2)

	result, err := client.apiRequest("GET", "/users.info?user="+userID, nil)
	if err != nil {
		return util.PushError(L, "get user failed: %v", err)
	}

	userTbl := L.NewTable()
	if user, ok := result["user"].(map[string]interface{}); ok {
		if id, ok := user["id"].(string); ok {
			userTbl.RawSetString("id", lua.LString(id))
		}
		if name, ok := user["name"].(string); ok {
			userTbl.RawSetString("name", lua.LString(name))
		}
		if realName, ok := user["real_name"].(string); ok {
			userTbl.RawSetString("real_name", lua.LString(realName))
		}
		if profile, ok := user["profile"].(map[string]interface{}); ok {
			if email, ok := profile["email"].(string); ok {
				userTbl.RawSetString("email", lua.LString(email))
			}
		}
	}

	return util.PushSuccess(L, userTbl)
}

func luaClientListUsers(L *lua.LState) int {
	client := checkSlackClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	opts := L.OptTable(2, nil)
	limit := ""
	if opts != nil {
		if v := L.GetField(opts, "limit"); v != lua.LNil {
			limit = fmt.Sprintf("&limit=%d", int(lua.LVAsNumber(v)))
		}
	}

	endpoint := "/users.list" + limit
	result, err := client.apiRequest("GET", endpoint, nil)
	if err != nil {
		return util.PushError(L, "list users failed: %v", err)
	}

	users := L.NewTable()
	if usrs, ok := result["members"].([]interface{}); ok {
		for i, u := range usrs {
			if userMap, ok := u.(map[string]interface{}); ok {
				userTbl := L.NewTable()
				if id, ok := userMap["id"].(string); ok {
					userTbl.RawSetString("id", lua.LString(id))
				}
				if name, ok := userMap["name"].(string); ok {
					userTbl.RawSetString("name", lua.LString(name))
				}
				if realName, ok := userMap["real_name"].(string); ok {
					userTbl.RawSetString("real_name", lua.LString(realName))
				}
				if deleted, ok := userMap["deleted"].(bool); ok {
					userTbl.RawSetString("deleted", lua.LBool(deleted))
				}
				if isBot, ok := userMap["is_bot"].(bool); ok {
					userTbl.RawSetString("is_bot", lua.LBool(isBot))
				}
				if profile, ok := userMap["profile"].(map[string]interface{}); ok {
					if email, ok := profile["email"].(string); ok {
						userTbl.RawSetString("email", lua.LString(email))
					}
					if displayName, ok := profile["display_name"].(string); ok {
						userTbl.RawSetString("display_name", lua.LString(displayName))
					}
				}
				users.RawSetInt(i+1, userTbl)
			}
		}
	}

	return util.PushSuccess(L, users)
}
