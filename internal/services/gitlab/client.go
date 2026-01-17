package gitlab

import (
	"context"
	"fmt"
	"net/url"

	"github.com/zepzeper/vulgar/internal/config"
	"github.com/zepzeper/vulgar/internal/httpclient"
)

const (
	DefaultBaseURL = "https://gitlab.com"
)

type Client struct {
	http     *httpclient.Client
	baseURL  string
	projects []string
}

type ClientOptions struct {
	Token    string
	URL      string
	Projects []string
}

func NewClient(opts ClientOptions) (*Client, error) {
	if opts.Token == "" {
		return nil, fmt.Errorf("gitlab token is required")
	}

	baseURL := opts.URL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	apiURL := baseURL
	if apiURL[len(apiURL)-1] == '/' {
		apiURL = apiURL + "api/v4"
	} else {
		apiURL = apiURL + "/api/v4"
	}

	httpClient := httpclient.New(
		httpclient.WithBaseURL(apiURL),
		httpclient.WithHeader("PRIVATE-TOKEN", opts.Token),
		httpclient.WithHeader("Content-Type", "application/json"),
		httpclient.WithRetry(2),
	)

	return &Client{
		http:     httpClient,
		baseURL:  baseURL,
		projects: opts.Projects,
	}, nil
}

func NewClientFromConfig() (*Client, error) {
	token, ok := config.GetGitLabToken()
	if !ok {
		return nil, fmt.Errorf("gitlab token not configured: run 'vulgar init' and set token in %s", config.ConfigPath())
	}

	return NewClient(ClientOptions{
		Token:    token,
		URL:      config.GetGitLabURL(),
		Projects: config.GetGitLabProjects(),
	})
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) Projects() []string {
	return c.projects
}

func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	resp, err := c.http.Get(ctx, "/user")
	if err != nil {
		return nil, fmt.Errorf("get user failed: %w", err)
	}

	if err := resp.CheckGitLab(); err != nil {
		return nil, err
	}

	var user User
	if err := resp.JSON(&user); err != nil {
		return nil, fmt.Errorf("parse user failed: %w", err)
	}

	return &user, nil
}

func (c *Client) GetProject(ctx context.Context, projectPath string) (*Project, error) {
	endpoint := fmt.Sprintf("/projects/%s", url.PathEscape(projectPath))

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get project failed: %w", err)
	}

	if err := resp.CheckGitLab(); err != nil {
		return nil, err
	}

	var project Project
	if err := resp.JSON(&project); err != nil {
		return nil, fmt.Errorf("parse project failed: %w", err)
	}

	return &project, nil
}

func (c *Client) ListProjects(ctx context.Context, opts ProjectListOptions) ([]Project, error) {
	q := opts.ToQuery()
	if q.Get("order_by") == "" {
		q.Set("order_by", "updated_at")
	}

	endpoint := "/projects?" + q.Encode()

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list projects failed: %w", err)
	}

	if err := resp.CheckGitLab(); err != nil {
		return nil, err
	}

	var projects []Project
	if err := resp.JSON(&projects); err != nil {
		return nil, fmt.Errorf("parse projects failed: %w", err)
	}

	return projects, nil
}

func (c *Client) ListCommits(ctx context.Context, projectPath string, opts CommitListOptions) ([]Commit, error) {
	q := opts.ToQuery()
	if q.Get("per_page") == "" {
		q.Set("per_page", "20")
	}

	endpoint := fmt.Sprintf("/projects/%s/repository/commits?%s", url.PathEscape(projectPath), q.Encode())

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list commits failed: %w", err)
	}

	if err := resp.CheckGitLab(); err != nil {
		return nil, err
	}

	var commits []Commit
	if err := resp.JSON(&commits); err != nil {
		return nil, fmt.Errorf("parse commits failed: %w", err)
	}

	return commits, nil
}

func (c *Client) ListMergeRequests(ctx context.Context, projectPath string, opts MergeRequestListOptions) ([]MergeRequest, error) {
	q := opts.ToQuery()
	if q.Get("per_page") == "" {
		q.Set("per_page", "20")
	}
	if q.Get("order_by") == "" {
		q.Set("order_by", "updated_at")
	}

	endpoint := fmt.Sprintf("/projects/%s/merge_requests?%s", url.PathEscape(projectPath), q.Encode())

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list merge requests failed: %w", err)
	}

	if err := resp.CheckGitLab(); err != nil {
		return nil, err
	}

	var mrs []MergeRequest
	if err := resp.JSON(&mrs); err != nil {
		return nil, fmt.Errorf("parse merge requests failed: %w", err)
	}

	return mrs, nil
}

func (c *Client) ListIssues(ctx context.Context, projectPath string, opts IssueListOptions) ([]Issue, error) {
	q := opts.ToQuery()
	if q.Get("per_page") == "" {
		q.Set("per_page", "20")
	}
	if q.Get("order_by") == "" {
		q.Set("order_by", "updated_at")
	}

	endpoint := fmt.Sprintf("/projects/%s/issues?%s", url.PathEscape(projectPath), q.Encode())

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list issues failed: %w", err)
	}

	if err := resp.CheckGitLab(); err != nil {
		return nil, err
	}

	var issues []Issue
	if err := resp.JSON(&issues); err != nil {
		return nil, fmt.Errorf("parse issues failed: %w", err)
	}

	return issues, nil
}

func (c *Client) ListPipelines(ctx context.Context, projectPath string, opts PipelineListOptions) ([]Pipeline, error) {
	q := opts.ToQuery()
	if q.Get("per_page") == "" {
		q.Set("per_page", "10")
	}
	if q.Get("order_by") == "" {
		q.Set("order_by", "updated_at")
	}

	endpoint := fmt.Sprintf("/projects/%s/pipelines?%s", url.PathEscape(projectPath), q.Encode())

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list pipelines failed: %w", err)
	}

	if err := resp.CheckGitLab(); err != nil {
		return nil, err
	}

	var pipelines []Pipeline
	if err := resp.JSON(&pipelines); err != nil {
		return nil, fmt.Errorf("parse pipelines failed: %w", err)
	}

	return pipelines, nil
}

func (c *Client) ListBranches(ctx context.Context, projectPath string, opts ListOptions) ([]Branch, error) {
	q := opts.ToQuery()
	if q.Get("per_page") == "" {
		q.Set("per_page", "20")
	}

	endpoint := fmt.Sprintf("/projects/%s/repository/branches?%s", url.PathEscape(projectPath), q.Encode())

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list branches failed: %w", err)
	}

	if err := resp.CheckGitLab(); err != nil {
		return nil, err
	}

	var branches []Branch
	if err := resp.JSON(&branches); err != nil {
		return nil, fmt.Errorf("parse branches failed: %w", err)
	}

	return branches, nil
}

func (c *Client) ListTags(ctx context.Context, projectPath string, opts ListOptions) ([]Tag, error) {
	q := opts.ToQuery()
	if q.Get("per_page") == "" {
		q.Set("per_page", "20")
	}

	endpoint := fmt.Sprintf("/projects/%s/repository/tags?%s", url.PathEscape(projectPath), q.Encode())

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list tags failed: %w", err)
	}

	if err := resp.CheckGitLab(); err != nil {
		return nil, err
	}

	var tags []Tag
	if err := resp.JSON(&tags); err != nil {
		return nil, fmt.Errorf("parse tags failed: %w", err)
	}

	return tags, nil
}

func (c *Client) CreateIssue(ctx context.Context, projectPath string, req CreateIssueRequest) (*Issue, error) {
	endpoint := fmt.Sprintf("/projects/%s/issues", url.PathEscape(projectPath))

	resp, err := c.http.NewRequest("POST", endpoint).
		Context(ctx).
		BodyJSON(req).
		Do()
	if err != nil {
		return nil, fmt.Errorf("create issue failed: %w", err)
	}

	if err := resp.CheckGitLab(); err != nil {
		return nil, err
	}

	var issue Issue
	if err := resp.JSON(&issue); err != nil {
		return nil, fmt.Errorf("parse issue failed: %w", err)
	}

	return &issue, nil
}

func (c *Client) CreateMergeRequest(ctx context.Context, projectPath string, req CreateMergeRequestRequest) (*MergeRequest, error) {
	endpoint := fmt.Sprintf("/projects/%s/merge_requests", url.PathEscape(projectPath))

	resp, err := c.http.NewRequest("POST", endpoint).
		Context(ctx).
		BodyJSON(req).
		Do()
	if err != nil {
		return nil, fmt.Errorf("create merge request failed: %w", err)
	}

	if err := resp.CheckGitLab(); err != nil {
		return nil, err
	}

	var mr MergeRequest
	if err := resp.JSON(&mr); err != nil {
		return nil, fmt.Errorf("parse merge request failed: %w", err)
	}

	return &mr, nil
}
