package github

import (
	"fmt"
	"net/url"
)

type ListOptions struct {
	PerPage int
	Page    int
	Sort    string // created, updated, pushed, full_name
	Direction string // asc, desc
}

type RepoListOptions struct {
	ListOptions
	Visibility string // all, public, private
	Type       string // all, owner, member
}

type IssueListOptions struct {
	ListOptions
	State     string // open, closed, all
	Labels    string // comma-separated
	Since     string // ISO8601 timestamp
	Assignee  string
	Creator   string
}

type PullRequestListOptions struct {
	ListOptions
	State string // open, closed, all
	Head  string // Filter by head branch
	Base  string // Filter by base branch
}

type CommitListOptions struct {
	ListOptions
	SHA   string // SHA or branch to start listing from
	Path  string // Only commits containing this path
	Since string // ISO8601 timestamp
	Until string // ISO8601 timestamp
}

func (o *ListOptions) ToQuery() url.Values {
	q := url.Values{}

	if o.PerPage > 0 {
		q.Set("per_page", fmt.Sprintf("%d", o.PerPage))
	}
	if o.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", o.Page))
	}
	if o.Sort != "" {
		q.Set("sort", o.Sort)
	}
	if o.Direction != "" {
		q.Set("direction", o.Direction)
	}

	return q
}

func (o *RepoListOptions) ToQuery() url.Values {
	q := o.ListOptions.ToQuery()

	if o.Visibility != "" && o.Visibility != "all" {
		q.Set("visibility", o.Visibility)
	}
	if o.Type != "" {
		q.Set("type", o.Type)
	}

	return q
}

func (o *IssueListOptions) ToQuery() url.Values {
	q := o.ListOptions.ToQuery()

	if o.State != "" {
		q.Set("state", o.State)
	}
	if o.Labels != "" {
		q.Set("labels", o.Labels)
	}
	if o.Since != "" {
		q.Set("since", o.Since)
	}
	if o.Assignee != "" {
		q.Set("assignee", o.Assignee)
	}
	if o.Creator != "" {
		q.Set("creator", o.Creator)
	}

	return q
}

func (o *PullRequestListOptions) ToQuery() url.Values {
	q := o.ListOptions.ToQuery()

	if o.State != "" {
		q.Set("state", o.State)
	}
	if o.Head != "" {
		q.Set("head", o.Head)
	}
	if o.Base != "" {
		q.Set("base", o.Base)
	}

	return q
}

func (o *CommitListOptions) ToQuery() url.Values {
	q := o.ListOptions.ToQuery()

	if o.SHA != "" {
		q.Set("sha", o.SHA)
	}
	if o.Path != "" {
		q.Set("path", o.Path)
	}
	if o.Since != "" {
		q.Set("since", o.Since)
	}
	if o.Until != "" {
		q.Set("until", o.Until)
	}

	return q
}
