package gitlab

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	gitlab "github.com/zepzeper/vulgar/internal/services/gitlab"
)

// Usage: local mrs, err = client:list_merge_requests("group/project", {state = "opened"})
func luaListMergeRequests(L *lua.LState) int {
	client := checkGitLabClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	project := L.CheckString(2)
	luaOpts := L.OptTable(3, nil)

	opts := gitlab.MergeRequestListOptions{}
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

	mrs, err := client.svc.ListMergeRequests(context.Background(), project, opts)
	if err != nil {
		return util.PushError(L, "list merge requests failed: %v", err)
	}

	return util.PushSuccess(L, mergeRequestsToLua(L, mrs))
}

// Usage: local mr, err = client:create_merge_request("group/project", {title = "Feature", source_branch = "feature", target_branch = "main"})
func luaCreateMergeRequest(L *lua.LState) int {
	client := checkGitLabClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	project := L.CheckString(2)
	opts := L.CheckTable(3)

	req := gitlab.CreateMergeRequestRequest{}

	if v := L.GetField(opts, "title"); v != lua.LNil {
		req.Title = lua.LVAsString(v)
	} else {
		return util.PushError(L, "title is required")
	}
	if v := L.GetField(opts, "source_branch"); v != lua.LNil {
		req.SourceBranch = lua.LVAsString(v)
	} else {
		return util.PushError(L, "source_branch is required")
	}
	if v := L.GetField(opts, "target_branch"); v != lua.LNil {
		req.TargetBranch = lua.LVAsString(v)
	} else {
		return util.PushError(L, "target_branch is required")
	}
	if v := L.GetField(opts, "description"); v != lua.LNil {
		req.Description = lua.LVAsString(v)
	}

	mr, err := client.svc.CreateMergeRequest(context.Background(), project, req)
	if err != nil {
		return util.PushError(L, "create merge request failed: %v", err)
	}

	return util.PushSuccess(L, mergeRequestToLua(L, mr))
}

func luaClientListMergeRequests(L *lua.LState) int {
	return luaListMergeRequests(L)
}

func luaClientCreateMergeRequest(L *lua.LState) int {
	return luaCreateMergeRequest(L)
}

func mergeRequestToLua(L *lua.LState, mr *gitlab.MergeRequest) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("iid", lua.LNumber(mr.IID))
	tbl.RawSetString("title", lua.LString(mr.Title))
	tbl.RawSetString("state", lua.LString(mr.State))
	tbl.RawSetString("source_branch", lua.LString(mr.SourceBranch))
	tbl.RawSetString("target_branch", lua.LString(mr.TargetBranch))
	tbl.RawSetString("web_url", lua.LString(mr.WebURL))
	tbl.RawSetString("created_at", lua.LString(mr.CreatedAt))
	tbl.RawSetString("updated_at", lua.LString(mr.UpdatedAt))
	if mr.Author != nil {
		tbl.RawSetString("author", lua.LString(mr.Author.Username))
	}
	return tbl
}

func mergeRequestsToLua(L *lua.LState, mrs []gitlab.MergeRequest) *lua.LTable {
	tbl := L.NewTable()
	for i, mr := range mrs {
		m := mr // avoid closure issues
		tbl.RawSetInt(i+1, mergeRequestToLua(L, &m))
	}
	return tbl
}
