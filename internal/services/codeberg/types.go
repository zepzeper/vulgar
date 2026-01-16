package codeberg

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
	Stars         int    `json:"stars_count"`
	Forks         int    `json:"forks_count"`
	DefaultBranch string `json:"default_branch"`
	UpdatedAt     string `json:"updated_at"`
}

type Issue struct {
	ID        int    `json:"id"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	User      *User  `json:"user"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

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
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type Branch struct {
	Label string `json:"label"`
	Ref   string `json:"ref"`
	SHA   string `json:"sha"`
}

type User struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

type CreateIssueRequest struct {
	Title string `json:"title"`
	Body  string `json:"body,omitempty"`
}

type CreatePullRequestRequest struct {
	Title string `json:"title"`
	Body  string `json:"body,omitempty"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}
