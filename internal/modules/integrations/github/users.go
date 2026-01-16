package github

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// Usage: local user, err = github.get_user(client)
func luaGetUser(L *lua.LState) int {
	client := checkGitHubClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	user, err := client.svc.GetCurrentUser(context.Background())
	if err != nil {
		return util.PushError(L, "get user failed: %v", err)
	}

	return util.PushSuccess(L, userToLua(L, user))
}

// Usage: local limit, err = github.rate_limit(client)
func luaRateLimit(L *lua.LState) int {
	client := checkGitHubClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	limit, err := client.svc.GetRateLimit(context.Background())
	if err != nil {
		return util.PushError(L, "get rate limit failed: %v", err)
	}

	tbl := L.NewTable()
	tbl.RawSetString("limit", lua.LNumber(limit.Limit))
	tbl.RawSetString("remaining", lua.LNumber(limit.Remaining))
	tbl.RawSetString("reset", lua.LNumber(limit.Reset))

	return util.PushSuccess(L, tbl)
}

// Client method wrappers
func luaClientGetUser(L *lua.LState) int {
	return luaGetUser(L)
}

func luaClientRateLimit(L *lua.LState) int {
	return luaRateLimit(L)
}
