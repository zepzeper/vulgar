package github

// Repository represents a GitHub repository.
type Repository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	Private       bool   `json:"private"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	SSHURL        string `json:"ssh_url"`
	Language      string `json:"language"`
	Stars         int    `json:"stargazers_count"`
	Forks         int    `json:"forks_count"`
	DefaultBranch string `json:"default_branch"`
	UpdatedAt     string `json:"updated_at"`
	PushedAt      string `json:"pushed_at"`
}

// Issue represents a GitHub issue.
type Issue struct {
	ID        int     `json:"id"`
	Number    int     `json:"number"`
	Title     string  `json:"title"`
	Body      string  `json:"body"`
	State     string  `json:"state"`
	HTMLURL   string  `json:"html_url"`
	User      *User   `json:"user"`
	Assignees []User  `json:"assignees"`
	Labels    []Label `json:"labels"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	ClosedAt  string  `json:"closed_at"`
}

// PullRequest represents a GitHub pull request.
type PullRequest struct {
	ID        int     `json:"id"`
	Number    int     `json:"number"`
	Title     string  `json:"title"`
	Body      string  `json:"body"`
	State     string  `json:"state"`
	HTMLURL   string  `json:"html_url"`
	User      *User   `json:"user"`
	Head      *Branch `json:"head"`
	Base      *Branch `json:"base"`
	Merged    bool    `json:"merged"`
	MergedAt  string  `json:"merged_at"`
	MergedBy  *User   `json:"merged_by"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	ClosedAt  string  `json:"closed_at"`
}

// Branch represents a branch reference in a PR.
type Branch struct {
	Label string `json:"label"`
	Ref   string `json:"ref"`
	SHA   string `json:"sha"`
}

// User represents a GitHub user.
type User struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Bio       string `json:"bio"`
}

// Label represents a GitHub label.
type Label struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Commit represents a GitHub commit.
type Commit struct {
	SHA     string      `json:"sha"`
	Message string      `json:"message"`
	Author  *GitAuthor  `json:"author"`
	HTMLURL string      `json:"html_url"`
	Commit  *CommitData `json:"commit"`
}

// CommitData contains the actual commit information.
type CommitData struct {
	Message string     `json:"message"`
	Author  *GitAuthor `json:"author"`
}

// GitAuthor represents a git author (name, email, date).
type GitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

// WorkflowRun represents a GitHub Actions workflow run.
type WorkflowRun struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HTMLURL    string `json:"html_url"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// RateLimit represents GitHub API rate limit info.
type RateLimit struct {
	Limit     int `json:"limit"`
	Remaining int `json:"remaining"`
	Reset     int `json:"reset"`
}

// CreateIssueRequest represents parameters for creating an issue.
type CreateIssueRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
}

// CreatePullRequestRequest represents parameters for creating a PR.
type CreatePullRequestRequest struct {
	Title string `json:"title"`
	Body  string `json:"body,omitempty"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}
