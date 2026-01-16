package gitlab

import (
	"sync"

	gitlab "github.com/zepzeper/vulgar/internal/services/gitlab"
)

const ModuleName = "integrations.gitlab"
const luaGitLabClientTypeName = "gitlab_client"

type gitlabClient struct {
	svc    *gitlab.Client
	mu     sync.Mutex
	closed bool
}
