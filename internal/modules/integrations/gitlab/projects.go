package gitlab

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	gitlab "github.com/zepzeper/vulgar/internal/services/gitlab"
)

// Usage: local project, err = client:get_project("group/project")
func luaGetProject(L *lua.LState) int {
	client := checkGitLabClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	project := L.CheckString(2)

	proj, err := client.svc.GetProject(context.Background(), project)
	if err != nil {
		return util.PushError(L, "get project failed: %v", err)
	}

	return util.PushSuccess(L, projectToLua(L, proj))
}

// Usage: local projects, err = client:list_projects({membership = true})
func luaListProjects(L *lua.LState) int {
	client := checkGitLabClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	luaOpts := L.OptTable(2, nil)

	opts := gitlab.ProjectListOptions{}
	if luaOpts != nil {
		if v := L.GetField(luaOpts, "membership"); v == lua.LTrue {
			opts.Membership = true
		}
		if v := L.GetField(luaOpts, "owned"); v == lua.LTrue {
			opts.Owned = true
		}
		if v := L.GetField(luaOpts, "search"); v != lua.LNil {
			opts.Search = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "per_page"); v != lua.LNil {
			opts.PerPage = int(lua.LVAsNumber(v))
		}
	}

	projects, err := client.svc.ListProjects(context.Background(), opts)
	if err != nil {
		return util.PushError(L, "list projects failed: %v", err)
	}

	return util.PushSuccess(L, projectsToLua(L, projects))
}

func luaClientGetProject(L *lua.LState) int {
	return luaGetProject(L)
}

func luaClientListProjects(L *lua.LState) int {
	return luaListProjects(L)
}

func projectToLua(L *lua.LState, p *gitlab.Project) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("id", lua.LNumber(p.ID))
	tbl.RawSetString("name", lua.LString(p.Name))
	tbl.RawSetString("path_with_namespace", lua.LString(p.PathWithNamespace))
	tbl.RawSetString("description", lua.LString(p.Description))
	tbl.RawSetString("web_url", lua.LString(p.WebURL))
	tbl.RawSetString("default_branch", lua.LString(p.DefaultBranch))
	tbl.RawSetString("visibility", lua.LString(p.Visibility))
	tbl.RawSetString("star_count", lua.LNumber(p.StarCount))
	tbl.RawSetString("forks_count", lua.LNumber(p.ForksCount))
	return tbl
}

func projectsToLua(L *lua.LState, projects []gitlab.Project) *lua.LTable {
	tbl := L.NewTable()
	for i, p := range projects {
		proj := p // avoid closure issues
		tbl.RawSetInt(i+1, projectToLua(L, &proj))
	}
	return tbl
}
