package github

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	github "github.com/zepzeper/vulgar/internal/services/github"
)

// Usage: local commits, err = github.list_commits(client, "owner", "repo", {sha = "main"})
func luaListCommits(L *lua.LState) int {
	client := checkGitHubClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	owner := L.CheckString(2)
	repo := L.CheckString(3)
	luaOpts := L.OptTable(4, nil)

	opts := github.CommitListOptions{}
	if luaOpts != nil {
		if v := L.GetField(luaOpts, "sha"); v != lua.LNil {
			opts.SHA = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "path"); v != lua.LNil {
			opts.Path = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "since"); v != lua.LNil {
			opts.Since = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "per_page"); v != lua.LNil {
			opts.PerPage = int(lua.LVAsNumber(v))
		}
	}

	commits, err := client.svc.ListCommits(context.Background(), owner, repo, opts)
	if err != nil {
		return util.PushError(L, "list commits failed: %v", err)
	}

	return util.PushSuccess(L, commitsToLua(L, commits))
}

// Client method wrapper
func luaClientListCommits(L *lua.LState) int {
	return luaListCommits(L)
}
