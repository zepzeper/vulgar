package github

import (
	"sync"

	github "github.com/zepzeper/vulgar/internal/services/github"
)

const ModuleName = "integrations.github"
const luaGitHubClientTypeName = "github_client"

type githubClient struct {
	svc    *github.Client
	mu     sync.Mutex
	closed bool
}
