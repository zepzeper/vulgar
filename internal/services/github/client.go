package github

import (
	"context"
	"fmt"

	"github.com/zepzeper/vulgar/internal/config"
	"github.com/zepzeper/vulgar/internal/httpclient"
)

const (
	DefaultBaseURL = "https://api.github.com"
)

type Client struct {
	http         *httpclient.Client
	defaultOwner string
}

type ClientOptions struct {
	// Token is the GitHub personal access token.
	Token string

	// DefaultOwner is the default owner for operations.
	DefaultOwner string
}

func NewClient(opts ClientOptions) (*Client, error) {
	if opts.Token == "" {
		return nil, fmt.Errorf("github token is required")
	}

	httpClient := httpclient.New(
		httpclient.WithBaseURL(DefaultBaseURL),
		httpclient.WithGitHubAuth(opts.Token),
		httpclient.WithRetry(2),
		httpclient.WithRateLimit(5),
	)

	return &Client{
		http:         httpClient,
		defaultOwner: opts.DefaultOwner,
	}, nil
}

func NewClientFromConfig() (*Client, error) {
	token, ok := config.GetGitHubToken()
	if !ok {
		return nil, fmt.Errorf("github token not configured: run 'vulgar init' and set token in %s", config.ConfigPath())
	}

	cfg := config.Get()

	return NewClient(ClientOptions{
		Token:        token,
		DefaultOwner: cfg.GitHub.DefaultOwner,
	})
}

func (c *Client) DefaultOwner() string {
	return c.defaultOwner
}

func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	resp, err := c.http.Get(ctx, "/user")
	if err != nil {
		return nil, fmt.Errorf("get user failed: %w", err)
	}

	if err := resp.CheckGitHub(); err != nil {
		return nil, err
	}

	var user User
	if err := resp.JSON(&user); err != nil {
		return nil, fmt.Errorf("parse user failed: %w", err)
	}

	return &user, nil
}

func (c *Client) GetRateLimit(ctx context.Context) (*RateLimit, error) {
	resp, err := c.http.Get(ctx, "/rate_limit")
	if err != nil {
		return nil, fmt.Errorf("get rate limit failed: %w", err)
	}

	if err := resp.CheckGitHub(); err != nil {
		return nil, err
	}

	var result struct {
		Resources struct {
			Core *RateLimit `json:"core"`
		} `json:"resources"`
	}
	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("parse rate limit failed: %w", err)
	}

	return result.Resources.Core, nil
}

func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s", owner, repo)

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get repository failed: %w", err)
	}

	if err := resp.CheckGitHub(); err != nil {
		return nil, err
	}

	var repository Repository
	if err := resp.JSON(&repository); err != nil {
		return nil, fmt.Errorf("parse repository failed: %w", err)
	}

	return &repository, nil
}

func (c *Client) ListUserRepositories(ctx context.Context, opts RepoListOptions) ([]Repository, error) {
	q := opts.ToQuery()
	if q.Get("sort") == "" {
		q.Set("sort", "updated")
	}

	endpoint := "/user/repos?" + q.Encode()

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list repositories failed: %w", err)
	}

	if err := resp.CheckGitHub(); err != nil {
		return nil, err
	}

	var repos []Repository
	if err := resp.JSON(&repos); err != nil {
		return nil, fmt.Errorf("parse repositories failed: %w", err)
	}

	return repos, nil
}

func (c *Client) ListOwnerRepositories(ctx context.Context, owner string, opts RepoListOptions) ([]Repository, error) {
	q := opts.ToQuery()
	if q.Get("sort") == "" {
		q.Set("sort", "updated")
	}

	endpoint := fmt.Sprintf("/users/%s/repos?%s", owner, q.Encode())

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list repositories failed: %w", err)
	}

	if err := resp.CheckGitHub(); err != nil {
		return nil, err
	}

	var repos []Repository
	if err := resp.JSON(&repos); err != nil {
		return nil, fmt.Errorf("parse repositories failed: %w", err)
	}

	return repos, nil
}

func (c *Client) ListIssues(ctx context.Context, owner, repo string, opts IssueListOptions) ([]Issue, error) {
	q := opts.ToQuery()
	if q.Get("per_page") == "" {
		q.Set("per_page", "20")
	}

	endpoint := fmt.Sprintf("/repos/%s/%s/issues?%s", owner, repo, q.Encode())

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list issues failed: %w", err)
	}

	if err := resp.CheckGitHub(); err != nil {
		return nil, err
	}

	var issues []Issue
	if err := resp.JSON(&issues); err != nil {
		return nil, fmt.Errorf("parse issues failed: %w", err)
	}

	return issues, nil
}

func (c *Client) ListPullRequests(ctx context.Context, owner, repo string, opts PullRequestListOptions) ([]PullRequest, error) {
	q := opts.ToQuery()
	if q.Get("per_page") == "" {
		q.Set("per_page", "20")
	}

	endpoint := fmt.Sprintf("/repos/%s/%s/pulls?%s", owner, repo, q.Encode())

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list pull requests failed: %w", err)
	}

	if err := resp.CheckGitHub(); err != nil {
		return nil, err
	}

	var prs []PullRequest
	if err := resp.JSON(&prs); err != nil {
		return nil, fmt.Errorf("parse pull requests failed: %w", err)
	}

	return prs, nil
}

func (c *Client) ListCommits(ctx context.Context, owner, repo string, opts CommitListOptions) ([]Commit, error) {
	q := opts.ToQuery()
	if q.Get("per_page") == "" {
		q.Set("per_page", "20")
	}

	endpoint := fmt.Sprintf("/repos/%s/%s/commits?%s", owner, repo, q.Encode())

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list commits failed: %w", err)
	}

	if err := resp.CheckGitHub(); err != nil {
		return nil, err
	}

	var commits []Commit
	if err := resp.JSON(&commits); err != nil {
		return nil, fmt.Errorf("parse commits failed: %w", err)
	}

	return commits, nil
}

func (c *Client) CreateIssue(ctx context.Context, owner, repo string, req CreateIssueRequest) (*Issue, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)

	resp, err := c.http.NewRequest("POST", endpoint).
		Context(ctx).
		BodyJSON(req).
		Do()
	if err != nil {
		return nil, fmt.Errorf("create issue failed: %w", err)
	}

	if err := resp.CheckGitHub(); err != nil {
		return nil, err
	}

	var issue Issue
	if err := resp.JSON(&issue); err != nil {
		return nil, fmt.Errorf("parse issue failed: %w", err)
	}

	return &issue, nil
}

func (c *Client) CreatePullRequest(ctx context.Context, owner, repo string, req CreatePullRequestRequest) (*PullRequest, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/pulls", owner, repo)

	resp, err := c.http.NewRequest("POST", endpoint).
		Context(ctx).
		BodyJSON(req).
		Do()
	if err != nil {
		return nil, fmt.Errorf("create pull request failed: %w", err)
	}

	if err := resp.CheckGitHub(); err != nil {
		return nil, err
	}

	var pr PullRequest
	if err := resp.JSON(&pr); err != nil {
		return nil, fmt.Errorf("parse pull request failed: %w", err)
	}

	return &pr, nil
}
