package gitlab

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	gitlab "github.com/zepzeper/vulgar/internal/services/gitlab"
)

// Usage: local issues, err = client:list_issues("group/project", {state = "opened"})
func luaListIssues(L *lua.LState) int {
	client := checkGitLabClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	project := L.CheckString(2)
	luaOpts := L.OptTable(3, nil)

	opts := gitlab.IssueListOptions{}
	if luaOpts != nil {
		if v := L.GetField(luaOpts, "state"); v != lua.LNil {
			opts.State = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "updated_after"); v != lua.LNil {
			opts.UpdatedAfter = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "per_page"); v != lua.LNil {
			opts.PerPage = int(lua.LVAsNumber(v))
		}
	}

	issues, err := client.svc.ListIssues(context.Background(), project, opts)
	if err != nil {
		return util.PushError(L, "list issues failed: %v", err)
	}

	return util.PushSuccess(L, issuesToLua(L, issues))
}

// Usage: local issue, err = client:create_issue("group/project", {title = "Bug", description = "..."})
func luaCreateIssue(L *lua.LState) int {
	client := checkGitLabClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	project := L.CheckString(2)
	opts := L.CheckTable(3)

	req := gitlab.CreateIssueRequest{}

	if v := L.GetField(opts, "title"); v != lua.LNil {
		req.Title = lua.LVAsString(v)
	} else {
		return util.PushError(L, "title is required")
	}

	if v := L.GetField(opts, "description"); v != lua.LNil {
		req.Description = lua.LVAsString(v)
	}
	if v := L.GetField(opts, "labels"); v != lua.LNil {
		req.Labels = lua.LVAsString(v) // comma-separated
	}
	if v := L.GetField(opts, "assignee_ids"); v != lua.LNil {
		if tbl, ok := v.(*lua.LTable); ok {
			tbl.ForEach(func(_, val lua.LValue) {
				req.AssigneeIDs = append(req.AssigneeIDs, int(lua.LVAsNumber(val)))
			})
		}
	}

	issue, err := client.svc.CreateIssue(context.Background(), project, req)
	if err != nil {
		return util.PushError(L, "create issue failed: %v", err)
	}

	return util.PushSuccess(L, issueToLua(L, issue))
}

func luaClientListIssues(L *lua.LState) int {
	return luaListIssues(L)
}

func luaClientCreateIssue(L *lua.LState) int {
	return luaCreateIssue(L)
}

func issueToLua(L *lua.LState, issue *gitlab.Issue) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("iid", lua.LNumber(issue.IID))
	tbl.RawSetString("title", lua.LString(issue.Title))
	tbl.RawSetString("state", lua.LString(issue.State))
	tbl.RawSetString("web_url", lua.LString(issue.WebURL))
	tbl.RawSetString("created_at", lua.LString(issue.CreatedAt))
	tbl.RawSetString("updated_at", lua.LString(issue.UpdatedAt))
	if issue.Author != nil {
		tbl.RawSetString("author", lua.LString(issue.Author.Username))
	}
	if len(issue.Labels) > 0 {
		labelsTbl := L.NewTable()
		for j, lbl := range issue.Labels {
			labelsTbl.RawSetInt(j+1, lua.LString(lbl))
		}
		tbl.RawSetString("labels", labelsTbl)
	}
	return tbl
}

func issuesToLua(L *lua.LState, issues []gitlab.Issue) *lua.LTable {
	tbl := L.NewTable()
	for i, issue := range issues {
		is := issue // avoid closure issues
		tbl.RawSetInt(i+1, issueToLua(L, &is))
	}
	return tbl
}
