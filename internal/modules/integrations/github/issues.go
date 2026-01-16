package github

import (
	"context"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	github "github.com/zepzeper/vulgar/internal/services/github"
)

// Usage: local issues, err = github.list_issues(client, "owner", "repo", {state = "open"})
func luaListIssues(L *lua.LState) int {
	client := checkGitHubClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	owner := L.CheckString(2)
	repo := L.CheckString(3)
	luaOpts := L.OptTable(4, nil)

	opts := github.IssueListOptions{}
	if luaOpts != nil {
		if v := L.GetField(luaOpts, "state"); v != lua.LNil {
			opts.State = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "labels"); v != lua.LNil {
			opts.Labels = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "per_page"); v != lua.LNil {
			opts.PerPage = int(lua.LVAsNumber(v))
		}
	}

	issues, err := client.svc.ListIssues(context.Background(), owner, repo, opts)
	if err != nil {
		return util.PushError(L, "list issues failed: %v", err)
	}

	// Filter out PRs (GitHub returns PRs in issues endpoint)
	filteredIssues := make([]github.Issue, 0)
	for _, i := range issues {
		if !strings.Contains(i.HTMLURL, "/pull/") {
			filteredIssues = append(filteredIssues, i)
		}
	}

	return util.PushSuccess(L, issuesToLua(L, filteredIssues))
}

// Usage: local issue, err = github.create_issue(client, "owner", "repo", {title = "Bug", body = "..."})
func luaCreateIssue(L *lua.LState) int {
	client := checkGitHubClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	owner := L.CheckString(2)
	repo := L.CheckString(3)
	opts := L.CheckTable(4)

	req := github.CreateIssueRequest{}

	if v := L.GetField(opts, "title"); v != lua.LNil {
		req.Title = lua.LVAsString(v)
	} else {
		return util.PushError(L, "title is required")
	}

	if v := L.GetField(opts, "body"); v != lua.LNil {
		req.Body = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "labels"); v != lua.LNil {
		if tbl, ok := v.(*lua.LTable); ok {
			tbl.ForEach(func(_, val lua.LValue) {
				req.Labels = append(req.Labels, lua.LVAsString(val))
			})
		}
	}
	if v := L.GetField(opts, "assignees"); v != lua.LNil {
		if tbl, ok := v.(*lua.LTable); ok {
			tbl.ForEach(func(_, val lua.LValue) {
				req.Assignees = append(req.Assignees, lua.LVAsString(val))
			})
		}
	}

	issue, err := client.svc.CreateIssue(context.Background(), owner, repo, req)
	if err != nil {
		return util.PushError(L, "create issue failed: %v", err)
	}

	return util.PushSuccess(L, issueToLua(L, issue))
}

// Client method wrappers
func luaClientListIssues(L *lua.LState) int {
	return luaListIssues(L)
}

func luaClientCreateIssue(L *lua.LState) int {
	return luaCreateIssue(L)
}
