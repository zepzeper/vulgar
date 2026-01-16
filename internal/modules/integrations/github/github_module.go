package github

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
)

var exports = map[string]lua.LGFunction{
	"client":       luaClient,
	"get_repo":     luaGetRepo,
	"list_repos":   luaListRepos,
	"list_issues":  luaListIssues,
	"list_prs":     luaListPRs,
	"list_commits": luaListCommits,
	"create_issue": luaCreateIssue,
	"create_pr":    luaCreatePR,
	"get_user":     luaGetUser,
	"rate_limit":   luaRateLimit,
}

func Loader(L *lua.LState) int {
	registerGitHubClientType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
