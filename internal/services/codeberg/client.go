package codeberg

import (
	"context"
	"fmt"
	"strings"

	"github.com/zepzeper/vulgar/internal/config"
	"github.com/zepzeper/vulgar/internal/httpclient"
)

const (
	DefaultBaseURL = "https://codeberg.org/api/v1"
)

type Client struct {
	http    *httpclient.Client
	baseURL string
}

type ClientOptions struct {
	Token string
	URL   string
}

func NewClient(opts ClientOptions) (*Client, error) {
	if opts.Token == "" {
		return nil, fmt.Errorf("codeberg token is required")
	}

	apiURL := DefaultBaseURL
	if opts.URL != "" {
		apiURL = strings.TrimSuffix(opts.URL, "/") + "/api/v1"
	}

	httpClient := httpclient.New(
		httpclient.WithBaseURL(apiURL),
		httpclient.WithHeader("Authorization", "token "+opts.Token),
		httpclient.WithHeader("Content-Type", "application/json"),
		httpclient.WithRetry(2),
	)

	return &Client{
		http:    httpClient,
		baseURL: opts.URL,
	}, nil
}

func NewClientFromConfig() (*Client, error) {
	token, ok := config.GetCodebergToken()
	if !ok {
		return nil, fmt.Errorf("codeberg token not configured: run 'vulgar init' and set token in %s", config.ConfigPath())
	}

	return NewClient(ClientOptions{
		Token: token,
		URL:   config.GetCodebergURL(),
	})
}

func (c *Client) BaseURL() string {
	if c.baseURL != "" {
		return c.baseURL
	}
	return "https://codeberg.org"
}

func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	resp, err := c.http.Get(ctx, "/user")
	if err != nil {
		return nil, fmt.Errorf("get user failed: %w", err)
	}

	if err := resp.CheckStatus(); err != nil {
		return nil, err
	}

	var user User
	if err := resp.JSON(&user); err != nil {
		return nil, fmt.Errorf("parse user failed: %w", err)
	}

	// Normalize login field
	if user.Login == "" {
		user.Login = user.Username
	}

	return &user, nil
}

func (c *Client) ListUserRepositories(ctx context.Context, limit int) ([]Repository, error) {
	endpoint := fmt.Sprintf("/user/repos?limit=%d", limit)

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list repositories failed: %w", err)
	}

	if err := resp.CheckStatus(); err != nil {
		return nil, err
	}

	var repos []Repository
	if err := resp.JSON(&repos); err != nil {
		return nil, fmt.Errorf("parse repositories failed: %w", err)
	}

	return repos, nil
}

func (c *Client) ListOwnerRepositories(ctx context.Context, owner string, limit int) ([]Repository, error) {
	endpoint := fmt.Sprintf("/users/%s/repos?limit=%d", owner, limit)

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list repositories failed: %w", err)
	}

	if err := resp.CheckStatus(); err != nil {
		return nil, err
	}

	var repos []Repository
	if err := resp.JSON(&repos); err != nil {
		return nil, fmt.Errorf("parse repositories failed: %w", err)
	}

	return repos, nil
}

func (c *Client) ListIssues(ctx context.Context, owner, repo, state string, limit int) ([]Issue, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/issues?state=%s&limit=%d", owner, repo, state, limit)

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list issues failed: %w", err)
	}

	if err := resp.CheckStatus(); err != nil {
		return nil, err
	}

	var issues []Issue
	if err := resp.JSON(&issues); err != nil {
		return nil, fmt.Errorf("parse issues failed: %w", err)
	}

	for i := range issues {
		if issues[i].User != nil && issues[i].User.Login == "" {
			issues[i].User.Login = issues[i].User.Username
		}
	}

	return issues, nil
}

func (c *Client) ListPullRequests(ctx context.Context, owner, repo, state string, limit int) ([]PullRequest, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/pulls?state=%s&limit=%d", owner, repo, state, limit)

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list pull requests failed: %w", err)
	}

	if err := resp.CheckStatus(); err != nil {
		return nil, err
	}

	var prs []PullRequest
	if err := resp.JSON(&prs); err != nil {
		return nil, fmt.Errorf("parse pull requests failed: %w", err)
	}

	for i := range prs {
		if prs[i].User != nil && prs[i].User.Login == "" {
			prs[i].User.Login = prs[i].User.Username
		}
	}

	return prs, nil
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

	if err := resp.CheckStatus(); err != nil {
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

	if err := resp.CheckStatus(); err != nil {
		return nil, err
	}

	var pr PullRequest
	if err := resp.JSON(&pr); err != nil {
		return nil, fmt.Errorf("parse pull request failed: %w", err)
	}

	return &pr, nil
}
