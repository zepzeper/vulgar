package codeberg

import (
	"context"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
	codeberg "github.com/zepzeper/vulgar/internal/services/codeberg"
)

const ModuleName = "integrations.codeberg"
const luaCodebergClientTypeName = "codeberg_client"

type codebergClient struct {
	svc    *codeberg.Client
	mu     sync.Mutex
	closed bool
}

var clientMethods = map[string]lua.LGFunction{
	"list_repos":   luaClientListRepos,
	"list_issues":  luaClientListIssues,
	"list_prs":     luaClientListPRs,
	"create_issue": luaClientCreateIssue,
	"create_pr":    luaClientCreatePR,
	"get_user":     luaClientGetUser,
}

func registerCodebergClientType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaCodebergClientTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), clientMethods))
	L.SetField(mt, "__gc", L.NewFunction(codebergClientGC))
}

func checkCodebergClient(L *lua.LState, idx int) *codebergClient {
	ud := L.CheckUserData(idx)
	if v, ok := ud.Value.(*codebergClient); ok {
		return v
	}
	L.ArgError(idx, "codeberg_client expected")
	return nil
}

func codebergClientGC(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if client, ok := ud.Value.(*codebergClient); ok {
		client.close()
	}
	return 0
}

func (c *codebergClient) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
}

// Usage: local client, err = codeberg.client()
// Or: local client, err = codeberg.client({token = "..."})
func luaClient(L *lua.LState) int {
	opts := L.OptTable(1, nil)

	var svc *codeberg.Client
	var err error

	if opts != nil {
		token := ""
		url := ""

		if v := L.GetField(opts, "token"); v != lua.LNil {
			token = lua.LVAsString(v)
		}
		if v := L.GetField(opts, "url"); v != lua.LNil {
			url = lua.LVAsString(v)
		}

		if token != "" {
			svc, err = codeberg.NewClient(codeberg.ClientOptions{
				Token: token,
				URL:   url,
			})
		}
	}

	if svc == nil && err == nil {
		svc, err = codeberg.NewClientFromConfig()
	}

	if err != nil {
		return util.PushError(L, "failed to create client: %v", err)
	}

	client := &codebergClient{svc: svc}

	ud := L.NewUserData()
	ud.Value = client
	L.SetMetatable(ud, L.GetTypeMetatable(luaCodebergClientTypeName))

	return util.PushSuccess(L, ud)
}

// Usage: local repos, err = client:list_repos("owner", {limit = 20})
func luaListRepos(L *lua.LState) int {
	client := checkCodebergClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	owner := ""
	if L.Get(2) != lua.LNil {
		owner = L.CheckString(2)
	}

	limit := 20
	if opts := L.OptTable(3, nil); opts != nil {
		if v := L.GetField(opts, "limit"); v != lua.LNil {
			limit = int(lua.LVAsNumber(v))
		}
	}

	var repos []codeberg.Repository
	var err error

	if owner != "" {
		repos, err = client.svc.ListOwnerRepositories(context.Background(), owner, limit)
	} else {
		repos, err = client.svc.ListUserRepositories(context.Background(), limit)
	}

	if err != nil {
		return util.PushError(L, "list repos failed: %v", err)
	}

	return util.PushSuccess(L, reposToLua(L, repos))
}

// Usage: local issues, err = client:list_issues("owner", "repo", {state = "open"})
func luaListIssues(L *lua.LState) int {
	client := checkCodebergClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	repo := L.CheckString(2)
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return util.PushError(L, "invalid repo format, use owner/repo")
	}
	owner, repoName := parts[0], parts[1]

	state := "open"
	limit := 20
	if opts := L.OptTable(3, nil); opts != nil {
		if v := L.GetField(opts, "state"); v != lua.LNil {
			state = lua.LVAsString(v)
		}
		if v := L.GetField(opts, "limit"); v != lua.LNil {
			limit = int(lua.LVAsNumber(v))
		}
	}

	issues, err := client.svc.ListIssues(context.Background(), owner, repoName, state, limit)
	if err != nil {
		return util.PushError(L, "list issues failed: %v", err)
	}

	return util.PushSuccess(L, issuesToLua(L, issues))
}

// Usage: local prs, err = client:list_prs("owner/repo", {state = "open"})
func luaListPRs(L *lua.LState) int {
	client := checkCodebergClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	repo := L.CheckString(2)
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return util.PushError(L, "invalid repo format, use owner/repo")
	}
	owner, repoName := parts[0], parts[1]

	state := "open"
	limit := 20
	if opts := L.OptTable(3, nil); opts != nil {
		if v := L.GetField(opts, "state"); v != lua.LNil {
			state = lua.LVAsString(v)
		}
		if v := L.GetField(opts, "limit"); v != lua.LNil {
			limit = int(lua.LVAsNumber(v))
		}
	}

	prs, err := client.svc.ListPullRequests(context.Background(), owner, repoName, state, limit)
	if err != nil {
		return util.PushError(L, "list prs failed: %v", err)
	}

	return util.PushSuccess(L, prsToLua(L, prs))
}

// Usage: local issue, err = client:create_issue("owner/repo", {title = "Bug", body = "..."})
func luaCreateIssue(L *lua.LState) int {
	client := checkCodebergClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	repo := L.CheckString(2)
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return util.PushError(L, "invalid repo format, use owner/repo")
	}
	owner, repoName := parts[0], parts[1]

	opts := L.CheckTable(3)
	req := codeberg.CreateIssueRequest{}

	if v := L.GetField(opts, "title"); v != lua.LNil {
		req.Title = lua.LVAsString(v)
	} else {
		return util.PushError(L, "title is required")
	}
	if v := L.GetField(opts, "body"); v != lua.LNil {
		req.Body = lua.LVAsString(v)
	}

	issue, err := client.svc.CreateIssue(context.Background(), owner, repoName, req)
	if err != nil {
		return util.PushError(L, "create issue failed: %v", err)
	}

	return util.PushSuccess(L, issueToLua(L, issue))
}

// Usage: local pr, err = client:create_pr("owner/repo", {title = "...", head = "feature", base = "main"})
func luaCreatePR(L *lua.LState) int {
	client := checkCodebergClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	repo := L.CheckString(2)
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return util.PushError(L, "invalid repo format, use owner/repo")
	}
	owner, repoName := parts[0], parts[1]

	opts := L.CheckTable(3)
	req := codeberg.CreatePullRequestRequest{}

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

	pr, err := client.svc.CreatePullRequest(context.Background(), owner, repoName, req)
	if err != nil {
		return util.PushError(L, "create pr failed: %v", err)
	}

	return util.PushSuccess(L, prToLua(L, pr))
}

// Usage: local user, err = client:get_user()
func luaGetUser(L *lua.LState) int {
	client := checkCodebergClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	user, err := client.svc.GetCurrentUser(context.Background())
	if err != nil {
		return util.PushError(L, "get user failed: %v", err)
	}

	return util.PushSuccess(L, userToLua(L, user))
}

func luaClientListRepos(L *lua.LState) int   { return luaListRepos(L) }
func luaClientListIssues(L *lua.LState) int  { return luaListIssues(L) }
func luaClientListPRs(L *lua.LState) int     { return luaListPRs(L) }
func luaClientCreateIssue(L *lua.LState) int { return luaCreateIssue(L) }
func luaClientCreatePR(L *lua.LState) int    { return luaCreatePR(L) }
func luaClientGetUser(L *lua.LState) int     { return luaGetUser(L) }

func repoToLua(L *lua.LState, r *codeberg.Repository) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("id", lua.LNumber(r.ID))
	tbl.RawSetString("name", lua.LString(r.Name))
	tbl.RawSetString("full_name", lua.LString(r.FullName))
	tbl.RawSetString("description", lua.LString(r.Description))
	tbl.RawSetString("private", lua.LBool(r.Private))
	tbl.RawSetString("html_url", lua.LString(r.HTMLURL))
	tbl.RawSetString("clone_url", lua.LString(r.CloneURL))
	tbl.RawSetString("language", lua.LString(r.Language))
	tbl.RawSetString("stars_count", lua.LNumber(r.Stars))
	tbl.RawSetString("forks_count", lua.LNumber(r.Forks))
	return tbl
}

func reposToLua(L *lua.LState, repos []codeberg.Repository) *lua.LTable {
	tbl := L.NewTable()
	for i, r := range repos {
		repo := r
		tbl.RawSetInt(i+1, repoToLua(L, &repo))
	}
	return tbl
}

func issueToLua(L *lua.LState, i *codeberg.Issue) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("number", lua.LNumber(i.Number))
	tbl.RawSetString("title", lua.LString(i.Title))
	tbl.RawSetString("body", lua.LString(i.Body))
	tbl.RawSetString("state", lua.LString(i.State))
	tbl.RawSetString("html_url", lua.LString(i.HTMLURL))
	tbl.RawSetString("created_at", lua.LString(i.CreatedAt))
	tbl.RawSetString("updated_at", lua.LString(i.UpdatedAt))
	if i.User != nil {
		tbl.RawSetString("user", lua.LString(i.User.Login))
	}
	return tbl
}

func issuesToLua(L *lua.LState, issues []codeberg.Issue) *lua.LTable {
	tbl := L.NewTable()
	for i, issue := range issues {
		is := issue
		tbl.RawSetInt(i+1, issueToLua(L, &is))
	}
	return tbl
}

func prToLua(L *lua.LState, pr *codeberg.PullRequest) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("number", lua.LNumber(pr.Number))
	tbl.RawSetString("title", lua.LString(pr.Title))
	tbl.RawSetString("body", lua.LString(pr.Body))
	tbl.RawSetString("state", lua.LString(pr.State))
	tbl.RawSetString("html_url", lua.LString(pr.HTMLURL))
	tbl.RawSetString("merged", lua.LBool(pr.Merged))
	tbl.RawSetString("created_at", lua.LString(pr.CreatedAt))
	tbl.RawSetString("updated_at", lua.LString(pr.UpdatedAt))
	if pr.User != nil {
		tbl.RawSetString("user", lua.LString(pr.User.Login))
	}
	if pr.Head != nil {
		tbl.RawSetString("head", lua.LString(pr.Head.Ref))
	}
	if pr.Base != nil {
		tbl.RawSetString("base", lua.LString(pr.Base.Ref))
	}
	return tbl
}

func prsToLua(L *lua.LState, prs []codeberg.PullRequest) *lua.LTable {
	tbl := L.NewTable()
	for i, pr := range prs {
		p := pr
		tbl.RawSetInt(i+1, prToLua(L, &p))
	}
	return tbl
}

func userToLua(L *lua.LState, u *codeberg.User) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("id", lua.LNumber(u.ID))
	tbl.RawSetString("login", lua.LString(u.Login))
	tbl.RawSetString("full_name", lua.LString(u.FullName))
	tbl.RawSetString("email", lua.LString(u.Email))
	tbl.RawSetString("html_url", lua.LString(u.HTMLURL))
	return tbl
}

var exports = map[string]lua.LGFunction{
	"client":       luaClient,
	"list_repos":   luaListRepos,
	"list_issues":  luaListIssues,
	"list_prs":     luaListPRs,
	"create_issue": luaCreateIssue,
	"create_pr":    luaCreatePR,
	"get_user":     luaGetUser,
}

func Loader(L *lua.LState) int {
	registerCodebergClientType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
