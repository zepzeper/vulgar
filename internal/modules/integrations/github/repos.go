package github

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
	github "github.com/zepzeper/vulgar/internal/services/github"
)

// Usage: local repo, err = github.get_repo(client, "owner", "repo")
func luaGetRepo(L *lua.LState) int {
	client := checkGitHubClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	owner := L.CheckString(2)
	repo := L.CheckString(3)

	repository, err := client.svc.GetRepository(context.Background(), owner, repo)
	if err != nil {
		return util.PushError(L, "get repo failed: %v", err)
	}

	return util.PushSuccess(L, repoToLua(L, repository))
}

// Usage: local repos, err = github.list_repos(client, "owner", {visibility = "all"})
// Or: local repos, err = github.list_repos(client, nil, {visibility = "all"}) -- for current user
func luaListRepos(L *lua.LState) int {
	client := checkGitHubClient(L, 1)
	if client == nil {
		return util.PushError(L, "invalid client")
	}

	owner := ""
	if L.Get(2) != lua.LNil {
		owner = L.CheckString(2)
	}

	luaOpts := L.OptTable(3, nil)

	opts := github.RepoListOptions{}
	if luaOpts != nil {
		if v := L.GetField(luaOpts, "visibility"); v != lua.LNil {
			opts.Visibility = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "type"); v != lua.LNil {
			opts.Type = lua.LVAsString(v)
		}
		if v := L.GetField(luaOpts, "per_page"); v != lua.LNil {
			opts.PerPage = int(lua.LVAsNumber(v))
		}
		if v := L.GetField(luaOpts, "sort"); v != lua.LNil {
			opts.Sort = lua.LVAsString(v)
		}
	}

	var repos []github.Repository
	var err error

	if owner != "" {
		repos, err = client.svc.ListOwnerRepositories(context.Background(), owner, opts)
	} else {
		repos, err = client.svc.ListUserRepositories(context.Background(), opts)
	}

	if err != nil {
		return util.PushError(L, "list repos failed: %v", err)
	}

	return util.PushSuccess(L, reposToLua(L, repos))
}

// Client method wrappers
func luaClientGetRepo(L *lua.LState) int {
	return luaGetRepo(L)
}

func luaClientListRepos(L *lua.LState) int {
	return luaListRepos(L)
}
