package gitlab

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
)

var exports = map[string]lua.LGFunction{
	"client":               luaClient,
	"list_commits":         luaListCommits,
	"list_merge_requests":  luaListMergeRequests,
	"list_issues":          luaListIssues,
	"list_pipelines":       luaListPipelines,
	"get_project":          luaGetProject,
	"list_projects":        luaListProjects,
	"create_issue":         luaCreateIssue,
	"create_merge_request": luaCreateMergeRequest,
	"get_user":             luaGetUser,
	"since_hours":          luaSinceHours,
}

func Loader(L *lua.LState) int {
	registerGitLabClientType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
