package github

import (
	lua "github.com/yuin/gopher-lua"
	github "github.com/zepzeper/vulgar/internal/services/github"
)

func repoToLua(L *lua.LState, r *github.Repository) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("id", lua.LNumber(r.ID))
	tbl.RawSetString("name", lua.LString(r.Name))
	tbl.RawSetString("full_name", lua.LString(r.FullName))
	tbl.RawSetString("description", lua.LString(r.Description))
	tbl.RawSetString("private", lua.LBool(r.Private))
	tbl.RawSetString("html_url", lua.LString(r.HTMLURL))
	tbl.RawSetString("clone_url", lua.LString(r.CloneURL))
	tbl.RawSetString("language", lua.LString(r.Language))
	tbl.RawSetString("stargazers_count", lua.LNumber(r.Stars))
	tbl.RawSetString("forks_count", lua.LNumber(r.Forks))
	tbl.RawSetString("default_branch", lua.LString(r.DefaultBranch))
	return tbl
}

func reposToLua(L *lua.LState, repos []github.Repository) *lua.LTable {
	tbl := L.NewTable()
	for i, r := range repos {
		repo := r
		tbl.RawSetInt(i+1, repoToLua(L, &repo))
	}
	return tbl
}

func issueToLua(L *lua.LState, i *github.Issue) *lua.LTable {
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

func issuesToLua(L *lua.LState, issues []github.Issue) *lua.LTable {
	tbl := L.NewTable()
	for i, issue := range issues {
		is := issue
		tbl.RawSetInt(i+1, issueToLua(L, &is))
	}
	return tbl
}

func prToLua(L *lua.LState, pr *github.PullRequest) *lua.LTable {
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

func prsToLua(L *lua.LState, prs []github.PullRequest) *lua.LTable {
	tbl := L.NewTable()
	for i, pr := range prs {
		p := pr
		tbl.RawSetInt(i+1, prToLua(L, &p))
	}
	return tbl
}

func commitToLua(L *lua.LState, c *github.Commit) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("sha", lua.LString(c.SHA))
	tbl.RawSetString("html_url", lua.LString(c.HTMLURL))

	if c.Commit != nil {
		tbl.RawSetString("message", lua.LString(c.Commit.Message))
		if c.Commit.Author != nil {
			tbl.RawSetString("author_name", lua.LString(c.Commit.Author.Name))
			tbl.RawSetString("author_email", lua.LString(c.Commit.Author.Email))
			tbl.RawSetString("date", lua.LString(c.Commit.Author.Date))
		}
	}
	return tbl
}

func commitsToLua(L *lua.LState, commits []github.Commit) *lua.LTable {
	tbl := L.NewTable()
	for i, c := range commits {
		commit := c
		tbl.RawSetInt(i+1, commitToLua(L, &commit))
	}
	return tbl
}

func userToLua(L *lua.LState, u *github.User) *lua.LTable {
	tbl := L.NewTable()
	tbl.RawSetString("id", lua.LNumber(u.ID))
	tbl.RawSetString("login", lua.LString(u.Login))
	tbl.RawSetString("name", lua.LString(u.Name))
	tbl.RawSetString("email", lua.LString(u.Email))
	tbl.RawSetString("html_url", lua.LString(u.HTMLURL))
	tbl.RawSetString("bio", lua.LString(u.Bio))
	return tbl
}
