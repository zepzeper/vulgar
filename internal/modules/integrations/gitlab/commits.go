package gitlab

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	gitlab "github.com/zepzeper/vulgar/internal/services/gitlab"
)

// Usage: local commits, err = client:list_commits("group/project", {since = "2024-01-01T00:00:00Z"})
func luaListCommits(L *lua.LState) int {
	client := checkGitLabClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	project := L.CheckString(2)
	luaOpts := L.OptTable(3, nil)

	opts := gitlab.CommitListOptions{}
	if luaOpts != nil {
		if v := L.GetField(luaOpts, "since"); v != lua.LNil {
			opts.Since = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "until"); v != lua.LNil {
			opts.Until = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "ref_name"); v != lua.LNil {
			opts.RefName = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "per_page"); v != lua.LNil {
			opts.PerPage = int(lua.LVAsNumber(v))
		}
	}

	commits, err := client.svc.ListCommits(context.Background(), project, opts)
	if err != nil {
		return util.PushError(L, "list commits failed: %v", err)
	}

	return util.PushSuccess(L, commitsToLua(L, commits))
}

func luaClientListCommits(L *lua.LState) int {
	return luaListCommits(L)
}

func commitToLua(L *lua.LState, c *gitlab.Commit) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("id", lua.LString(c.ID))
	tbl.RawSetString("short_id", lua.LString(c.ShortID))
	tbl.RawSetString("title", lua.LString(c.Title))
	tbl.RawSetString("message", lua.LString(c.Message))
	tbl.RawSetString("author_name", lua.LString(c.AuthorName))
	tbl.RawSetString("author_email", lua.LString(c.AuthorEmail))
	tbl.RawSetString("created_at", lua.LString(c.CreatedAt))
	tbl.RawSetString("web_url", lua.LString(c.WebURL))
	return tbl
}

func commitsToLua(L *lua.LState, commits []gitlab.Commit) *lua.LTable {
	tbl := L.NewTable()
	for i, c := range commits {
		commit := c // avoid closure issues
		tbl.RawSetInt(i+1, commitToLua(L, &commit))
	}
	return tbl
}
