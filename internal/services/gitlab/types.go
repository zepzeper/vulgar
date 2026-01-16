package gitlab

type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
	Description       string `json:"description"`
	WebURL            string `json:"web_url"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	HTTPURLToRepo     string `json:"http_url_to_repo"`
	DefaultBranch     string `json:"default_branch"`
	Visibility        string `json:"visibility"`
	StarCount         int    `json:"star_count"`
	ForksCount        int    `json:"forks_count"`
	LastActivityAt    string `json:"last_activity_at"`
}

type Commit struct {
	ID             string   `json:"id"`
	ShortID        string   `json:"short_id"`
	Title          string   `json:"title"`
	Message        string   `json:"message"`
	AuthorName     string   `json:"author_name"`
	AuthorEmail    string   `json:"author_email"`
	CommitterName  string   `json:"committer_name"`
	CommitterEmail string   `json:"committer_email"`
	CreatedAt      string   `json:"created_at"`
	WebURL         string   `json:"web_url"`
	ParentIDs      []string `json:"parent_ids"`
}

type MergeRequest struct {
	ID           int      `json:"id"`
	IID          int      `json:"iid"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	State        string   `json:"state"`
	SourceBranch string   `json:"source_branch"`
	TargetBranch string   `json:"target_branch"`
	Author       *User    `json:"author"`
	Assignee     *User    `json:"assignee"`
	Assignees    []User   `json:"assignees"`
	Labels       []string `json:"labels"`
	WebURL       string   `json:"web_url"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	MergedAt     string   `json:"merged_at"`
	ClosedAt     string   `json:"closed_at"`
	MergedBy     *User    `json:"merged_by"`
}

type Issue struct {
	ID          int      `json:"id"`
	IID         int      `json:"iid"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	State       string   `json:"state"`
	Author      *User    `json:"author"`
	Assignee    *User    `json:"assignee"`
	Assignees   []User   `json:"assignees"`
	Labels      []string `json:"labels"`
	WebURL      string   `json:"web_url"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	ClosedAt    string   `json:"closed_at"`
	ClosedBy    *User    `json:"closed_by"`
}

type Pipeline struct {
	ID        int    `json:"id"`
	Status    string `json:"status"`
	Ref       string `json:"ref"`
	SHA       string `json:"sha"`
	WebURL    string `json:"web_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Source    string `json:"source"`
}

// User represents a GitLab user.
type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	State     string `json:"state"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
}

type Branch struct {
	Name      string  `json:"name"`
	Commit    *Commit `json:"commit"`
	Merged    bool    `json:"merged"`
	Protected bool    `json:"protected"`
	Default   bool    `json:"default"`
	WebURL    string  `json:"web_url"`
}

type Tag struct {
	Name      string  `json:"name"`
	Message   string  `json:"message"`
	Target    string  `json:"target"`
	Commit    *Commit `json:"commit"`
	Protected bool    `json:"protected"`
}

type CreateIssueRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Labels      string `json:"labels,omitempty"` // comma-separated
	AssigneeIDs []int  `json:"assignee_ids,omitempty"`
	MilestoneID int    `json:"milestone_id,omitempty"`
}

type CreateMergeRequestRequest struct {
	Title        string `json:"title"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Description  string `json:"description,omitempty"`
	Labels       string `json:"labels,omitempty"` // comma-separated
	AssigneeIDs  []int  `json:"assignee_ids,omitempty"`
}
