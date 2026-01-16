package gitlab

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	gitlab "github.com/zepzeper/vulgar/internal/services/gitlab"
)

var clientMethods = map[string]lua.LGFunction{
	"list_commits":         luaClientListCommits,
	"list_merge_requests":  luaClientListMergeRequests,
	"list_issues":          luaClientListIssues,
	"list_pipelines":       luaClientListPipelines,
	"get_project":          luaClientGetProject,
	"list_projects":        luaClientListProjects,
	"create_issue":         luaClientCreateIssue,
	"create_merge_request": luaClientCreateMergeRequest,
	"get_user":             luaClientGetUser,
	"config":               luaClientConfig,
}

func registerGitLabClientType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaGitLabClientTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), clientMethods))
	L.SetField(mt, "__gc", L.NewFunction(gitlabClientGC))
}

func checkGitLabClient(L *lua.LState, idx int) *gitlabClient {
	ud := L.CheckUserData(idx)
	if v, ok := ud.Value.(*gitlabClient); ok {
		return v
	}
	L.ArgError(idx, "gitlab_client expected")
	return nil
}

func gitlabClientGC(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if client, ok := ud.Value.(*gitlabClient); ok {
		client.close()
	}
	return 0
}

func (c *gitlabClient) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
}

func (c *gitlabClient) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

// Usage: local client, err = gitlab.client()  -- Uses config from ~/.config/vulgar/config.toml
// Or: local client, err = gitlab.client({token = "glpat-...", url = "https://gitlab.example.com"})
func luaClient(L *lua.LState) int {
	opts := L.OptTable(1, nil)

	var svc *gitlab.Client
	var err error

	// Try explicit options first
	if opts != nil {
		token := ""
		url := ""
		var projects []string

		if v := L.GetField(opts, "token"); v != lua.LNil {
			token = lua.LVAsString(v)
		}
		if v := L.GetField(opts, "url"); v != lua.LNil {
			url = lua.LVAsString(v)
		}
		if v := L.GetField(opts, "projects"); v != lua.LNil {
			if tbl, ok := v.(*lua.LTable); ok {
				tbl.ForEach(func(_, val lua.LValue) {
					if s, ok := val.(lua.LString); ok {
						projects = append(projects, string(s))
					}
				})
			}
		}

		if token != "" {
			// Use explicit options
			svc, err = gitlab.NewClient(gitlab.ClientOptions{
				Token:    token,
				URL:      url,
				Projects: projects,
			})
		}
	}

	// Fall back to config
	if svc == nil && err == nil {
		svc, err = gitlab.NewClientFromConfig()
	}

	if err != nil {
		return util.PushError(L, "failed to create client: %v", err)
	}

	client := &gitlabClient{svc: svc}

	ud := L.NewUserData()
	ud.Value = client
	L.SetMetatable(ud, L.GetTypeMetatable(luaGitLabClientTypeName))

	return util.PushSuccess(L, ud)
}

// Usage: local config = client:config()
// Returns: {url = "...", projects = {...}}
func luaClientConfig(L *lua.LState) int {
	client := checkGitLabClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	result := L.NewTable()
	result.RawSetString("url", lua.LString(client.svc.BaseURL()))

	projectsTbl := L.NewTable()
	for i, p := range client.svc.Projects() {
		projectsTbl.RawSetInt(i+1, lua.LString(p))
	}
	result.RawSetString("projects", projectsTbl)

	return util.PushSuccess(L, result)
}
