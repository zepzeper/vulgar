package github

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	github "github.com/zepzeper/vulgar/internal/services/github"
)

// Usage: local prs, err = github.list_prs(client, "owner", "repo", {state = "open"})
func luaListPRs(L *lua.LState) int {
	client := checkGitHubClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	owner := L.CheckString(2)
	repo := L.CheckString(3)
	luaOpts := L.OptTable(4, nil)

	opts := github.PullRequestListOptions{}
	if luaOpts != nil {
		if v := L.GetField(luaOpts, "state"); v != lua.LNil {
			opts.State = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "base"); v != lua.LNil {
			opts.Base = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "head"); v != lua.LNil {
			opts.Head = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "per_page"); v != lua.LNil {
			opts.PerPage = int(lua.LVAsNumber(v))
		}
	}

	prs, err := client.svc.ListPullRequests(context.Background(), owner, repo, opts)
	if err != nil {
		return util.PushError(L, "list prs failed: %v", err)
	}

	return util.PushSuccess(L, prsToLua(L, prs))
}

// Usage: local pr, err = github.create_pr(client, "owner", "repo", {title = "...", head = "feature", base = "main"})
func luaCreatePR(L *lua.LState) int {
	client := checkGitHubClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	owner := L.CheckString(2)
	repo := L.CheckString(3)
	opts := L.CheckTable(4)

	req := github.CreatePullRequestRequest{}

	if v := L.GetField(opts, "title"); v != lua.LNil {
		req.Title = lua.LVAsString(v)
	} else {
		return util.PushError(L, "title is required")
	}
	if v := L.GetField(opts, "head"); v != lua.LNil {
		req.Head = lua.LVAsString(v)
	} else {
		return util.PushError(L, "head is required")
	}
	if v := L.GetField(opts, "base"); v != lua.LNil {
		req.Base = lua.LVAsString(v)
	} else {
		return util.PushError(L, "base is required")
	}
	if v := L.GetField(opts, "body"); v != lua.LNil {
		req.Body = lua.LVAsString(v)
	}

	pr, err := client.svc.CreatePullRequest(context.Background(), owner, repo, req)
	if err != nil {
		return util.PushError(L, "create pr failed: %v", err)
	}

	return util.PushSuccess(L, prToLua(L, pr))
}

// Client method wrappers
func luaClientListPRs(L *lua.LState) int {
	return luaListPRs(L)
}

func luaClientCreatePR(L *lua.LState) int {
	return luaCreatePR(L)
}
