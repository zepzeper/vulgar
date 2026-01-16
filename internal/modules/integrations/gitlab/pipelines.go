package gitlab

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	gitlab "github.com/zepzeper/vulgar/internal/services/gitlab"
)

// Usage: local pipelines, err = client:list_pipelines("group/project", {status = "success"})
func luaListPipelines(L *lua.LState) int {
	client := checkGitLabClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	project := L.CheckString(2)
	luaOpts := L.OptTable(3, nil)

	opts := gitlab.PipelineListOptions{}
	if luaOpts != nil {
		if v := L.GetField(luaOpts, "status"); v != lua.LNil {
			opts.Status = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "ref"); v != lua.LNil {
			opts.Ref = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "updated_after"); v != lua.LNil {
			opts.UpdatedAfter = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "per_page"); v != lua.LNil {
			opts.PerPage = int(lua.LVAsNumber(v))
		}
	}

	pipelines, err := client.svc.ListPipelines(context.Background(), project, opts)
	if err != nil {
		return util.PushError(L, "list pipelines failed: %v", err)
	}

	return util.PushSuccess(L, pipelinesToLua(L, pipelines))
}

func luaClientListPipelines(L *lua.LState) int {
	return luaListPipelines(L)
}

func pipelineToLua(L *lua.LState, p *gitlab.Pipeline) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("id", lua.LNumber(p.ID))
	tbl.RawSetString("status", lua.LString(p.Status))
	tbl.RawSetString("ref", lua.LString(p.Ref))
	tbl.RawSetString("sha", lua.LString(p.SHA))
	tbl.RawSetString("web_url", lua.LString(p.WebURL))
	tbl.RawSetString("created_at", lua.LString(p.CreatedAt))
	tbl.RawSetString("updated_at", lua.LString(p.UpdatedAt))
	return tbl
}

func pipelinesToLua(L *lua.LState, pipelines []gitlab.Pipeline) *lua.LTable {
	tbl := L.NewTable()
	for i, p := range pipelines {
		pipeline := p // avoid closure issues
		tbl.RawSetInt(i+1, pipelineToLua(L, &pipeline))
	}
	return tbl
}
