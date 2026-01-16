package git

import (
	"context"
	"fmt"

	"github.com/zepzeper/vulgar/internal/cli"
	"github.com/zepzeper/vulgar/internal/httpclient"
)

const (
	GitHubAPIBase   = "https://api.github.com"
	GitLabAPIBase   = "https://gitlab.com/api/v4"
	CodebergAPIBase = "https://codeberg.org/api/v1"
)

type RepoInfo struct {
	ID          int
	Name        string
	FullName    string
	Description string
	Private     bool
	URL         string
	CloneURL    string
	Language    string
	Stars       int
	Forks       int
	UpdatedAt   string
}

type IssueInfo struct {
	Number    int
	Title     string
	State     string
	Author    string
	Assignees []string
	Labels    []string
	CreatedAt string
	UpdatedAt string
	URL       string
}

type PRInfo struct {
	Number       int
	Title        string
	State        string
	Author       string
	SourceBranch string
	TargetBranch string
	IsMerged     bool
	CreatedAt    string
	UpdatedAt    string
	URL          string
}

type UserInfo struct {
	ID        int
	Login     string
	Name      string
	Email     string
	AvatarURL string
	Bio       string
	URL       string
}

func formatTimeAgo(timeStr string) string {
	return cli.FormatTimeAgo(timeStr)
}

func PrintReposTable(repos []RepoInfo) {
	if len(repos) == 0 {
		cli.PrintWarning("No repositories found")
		return
	}

	columns := []cli.Column{
		{Title: "Name", Width: 30},
		{Title: "Description", Width: 40},
		{Title: "Lang", Width: 10},
		{Title: "Stars", Width: 6},
		{Title: "Updated", Width: 12},
	}

	rows := make([][]string, len(repos))
	for i, repo := range repos {
		visibility := ""
		if repo.Private {
			visibility = " [P]"
		}
		rows[i] = []string{
			cli.Truncate(repo.Name+visibility, 30),
			cli.Truncate(repo.Description, 40),
			cli.Truncate(repo.Language, 10),
			fmt.Sprintf("%d", repo.Stars),
			formatTimeAgo(repo.UpdatedAt),
		}
	}

	cli.PrintTable(columns, rows)
}

func PrintIssuesTable(issues []IssueInfo) {
	if len(issues) == 0 {
		cli.PrintWarning("No issues found")
		return
	}

	columns := []cli.Column{
		{Title: "#", Width: 6},
		{Title: "Title", Width: 45},
		{Title: "State", Width: 10},
		{Title: "Author", Width: 15},
		{Title: "Updated", Width: 12},
	}

	rows := make([][]string, len(issues))
	for i, issue := range issues {
		rows[i] = []string{
			fmt.Sprintf("%d", issue.Number),
			cli.Truncate(issue.Title, 45),
			issue.State,
			cli.Truncate(issue.Author, 15),
			formatTimeAgo(issue.UpdatedAt),
		}
	}

	cli.PrintTable(columns, rows)
}

func PrintPRsTable(prs []PRInfo) {
	if len(prs) == 0 {
		cli.PrintWarning("No pull requests found")
		return
	}

	columns := []cli.Column{
		{Title: "#", Width: 6},
		{Title: "Title", Width: 40},
		{Title: "State", Width: 10},
		{Title: "Author", Width: 12},
		{Title: "Branch", Width: 20},
	}

	rows := make([][]string, len(prs))
	for i, pr := range prs {
		state := pr.State
		if pr.IsMerged {
			state = "merged"
		}
		rows[i] = []string{
			fmt.Sprintf("%d", pr.Number),
			cli.Truncate(pr.Title, 40),
			state,
			cli.Truncate(pr.Author, 12),
			cli.Truncate(pr.SourceBranch, 20),
		}
	}

	cli.PrintTable(columns, rows)
}

func PrintUserInfo(user UserInfo, platform string) {
	cli.PrintHeader(fmt.Sprintf("%s User Info", platform))
	fmt.Printf("  Username: %s\n", cli.Code(user.Login))
	if user.Name != "" {
		fmt.Printf("  Name:     %s\n", user.Name)
	}
	if user.Email != "" {
		fmt.Printf("  Email:    %s\n", user.Email)
	}
	if user.Bio != "" {
		fmt.Printf("  Bio:      %s\n", cli.Truncate(user.Bio, 60))
	}
	fmt.Printf("  URL:      %s\n", user.URL)
}

func newGitHubClient(token string) *httpclient.Client {
	return httpclient.New(
		httpclient.WithBaseURL(GitHubAPIBase),
		httpclient.WithGitHubAuth(token),
		httpclient.WithRetry(2),
		httpclient.WithRateLimit(5),
	)
}

func newGitLabClient(baseURL, token string) *httpclient.Client {
	if baseURL == "" {
		baseURL = GitLabAPIBase
	}
	return httpclient.New(
		httpclient.WithBaseURL(baseURL),
		httpclient.WithGitLabAuth(token),
		httpclient.WithRetry(2),
		httpclient.WithRateLimit(5),
	)
}

func newCodebergClient(baseURL, token string) *httpclient.Client {
	if baseURL == "" {
		baseURL = CodebergAPIBase
	}
	return httpclient.New(
		httpclient.WithBaseURL(baseURL),
		httpclient.WithBearerToken(token),
		httpclient.WithHeader("Accept", "application/json"),
		httpclient.WithRetry(2),
		httpclient.WithRateLimit(5),
	)
}

func fetchJSONArray(client *httpclient.Client, endpoint string) ([]interface{}, error) {
	resp, err := client.Get(context.Background(), endpoint)
	if err != nil {
		return nil, err
	}
	if err := resp.CheckStatus(); err != nil {
		return nil, err
	}
	var result []interface{}
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}
