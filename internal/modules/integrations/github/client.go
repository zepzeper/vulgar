package github

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	github "github.com/zepzeper/vulgar/internal/services/github"
)

var clientMethods = map[string]lua.LGFunction{
	"get_repo":     luaClientGetRepo,
	"list_repos":   luaClientListRepos,
	"list_issues":  luaClientListIssues,
	"list_prs":     luaClientListPRs,
	"list_commits": luaClientListCommits,
	"create_issue": luaClientCreateIssue,
	"create_pr":    luaClientCreatePR,
	"get_user":     luaClientGetUser,
	"rate_limit":   luaClientRateLimit,
}

func registerGitHubClientType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaGitHubClientTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), clientMethods))
	L.SetField(mt, "__gc", L.NewFunction(githubClientGC))
}

func checkGitHubClient(L *lua.LState, idx int) *githubClient {
	ud := L.CheckUserData(idx)
	if v, ok := ud.Value.(*githubClient); ok {
		return v
	}
	L.ArgError(idx, "github_client expected")
	return nil
}

func githubClientGC(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if client, ok := ud.Value.(*githubClient); ok {
		client.close()
	}
	return 0
}

func (c *githubClient) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
}

// Usage: local client, err = github.client()
// Or: local client, err = github.client({token = "ghp_..."})
func luaClient(L *lua.LState) int {
	opts := L.OptTable(1, nil)

	var svc *github.Client
	var err error

	if opts != nil {
		token := ""
		defaultOwner := ""

		if v := L.GetField(opts, "token"); v != lua.LNil {
			token = lua.LVAsString(v)
		}
		if v := L.GetField(opts, "default_owner"); v != lua.LNil {
			defaultOwner = lua.LVAsString(v)
		}

		if token != "" {
			svc, err = github.NewClient(github.ClientOptions{
				Token:        token,
				DefaultOwner: defaultOwner,
			})
		}
	}

	if svc == nil && err == nil {
		svc, err = github.NewClientFromConfig()
	}

	if err != nil {
		return util.PushError(L, "failed to create client: %v", err)
	}

	client := &githubClient{svc: svc}

	ud := L.NewUserData()
	ud.Value = client
	L.SetMetatable(ud, L.GetTypeMetatable(luaGitHubClientTypeName))

	return util.PushSuccess(L, ud)
}
