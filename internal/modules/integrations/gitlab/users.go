package gitlab

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	gitlab "github.com/zepzeper/vulgar/internal/services/gitlab"
)

// Usage: local user, err = client:get_user()  -- current user
func luaGetUser(L *lua.LState) int {
	client := checkGitLabClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	user, err := client.svc.GetCurrentUser(context.Background())
	if err != nil {
		return util.PushError(L, "get user failed: %v", err)
	}

	return util.PushSuccess(L, userToLua(L, user))
}

func luaClientGetUser(L *lua.LState) int {
	return luaGetUser(L)
}

func userToLua(L *lua.LState, u *gitlab.User) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("id", lua.LNumber(u.ID))
	tbl.RawSetString("username", lua.LString(u.Username))
	tbl.RawSetString("name", lua.LString(u.Name))
	tbl.RawSetString("email", lua.LString(u.Email))
	tbl.RawSetString("web_url", lua.LString(u.WebURL))
	return tbl
}
