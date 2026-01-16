package gitlab

import (
	"fmt"
	"net/url"
	"time"
)

type ListOptions struct {
	PerPage int
	Page    int

	Since string // created_after or since
	Until string // created_before or until

	State string // opened, closed, merged, all

	OrderBy string // created_at, updated_at
	Sort    string // asc, desc
}

type ProjectListOptions struct {
	ListOptions

	Membership bool
	Owned      bool

	Search string

	Visibility string // public, private, internal
}

type CommitListOptions struct {
	ListOptions

	RefName string

	Path string

	Author string
}

type MergeRequestListOptions struct {
	ListOptions

	UpdatedAfter  string
	UpdatedBefore string

	AuthorID   int
	AssigneeID int

	Labels string // comma-separated

	Scope string // created_by_me, assigned_to_me, all
}

type IssueListOptions struct {
	ListOptions

	UpdatedAfter  string
	UpdatedBefore string

	AuthorID   int
	AssigneeID int

	Labels string // comma-separated

	Scope string // created_by_me, assigned_to_me, all
}

type PipelineListOptions struct {
	ListOptions

	UpdatedAfter  string
	UpdatedBefore string

	Status string // running, pending, success, failed, canceled, skipped

	Ref string

	Source string // push, web, trigger, schedule, api, pipeline, etc.
}

func (o *ListOptions) ToQuery() url.Values {
	q := url.Values{}

	if o.PerPage > 0 {
		q.Set("per_page", fmt.Sprintf("%d", o.PerPage))
	}
	if o.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", o.Page))
	}
	if o.Since != "" {
		q.Set("since", o.Since)
	}
	if o.Until != "" {
		q.Set("until", o.Until)
	}
	if o.State != "" {
		q.Set("state", o.State)
	}
	if o.OrderBy != "" {
		q.Set("order_by", o.OrderBy)
	}
	if o.Sort != "" {
		q.Set("sort", o.Sort)
	}

	return q
}

func (o *ProjectListOptions) ToQuery() url.Values {
	q := o.ListOptions.ToQuery()

	if o.Membership {
		q.Set("membership", "true")
	}
	if o.Owned {
		q.Set("owned", "true")
	}
	if o.Search != "" {
		q.Set("search", o.Search)
	}
	if o.Visibility != "" {
		q.Set("visibility", o.Visibility)
	}

	return q
}

func (o *CommitListOptions) ToQuery() url.Values {
	q := o.ListOptions.ToQuery()

	if o.RefName != "" {
		q.Set("ref_name", o.RefName)
	}
	if o.Path != "" {
		q.Set("path", o.Path)
	}
	if o.Author != "" {
		q.Set("author", o.Author)
	}

	return q
}

func (o *MergeRequestListOptions) ToQuery() url.Values {
	q := o.ListOptions.ToQuery()

	if o.UpdatedAfter != "" {
		q.Set("updated_after", o.UpdatedAfter)
	}
	if o.UpdatedBefore != "" {
		q.Set("updated_before", o.UpdatedBefore)
	}
	if o.AuthorID > 0 {
		q.Set("author_id", fmt.Sprintf("%d", o.AuthorID))
	}
	if o.AssigneeID > 0 {
		q.Set("assignee_id", fmt.Sprintf("%d", o.AssigneeID))
	}
	if o.Labels != "" {
		q.Set("labels", o.Labels)
	}
	if o.Scope != "" {
		q.Set("scope", o.Scope)
	}

	return q
}

func (o *IssueListOptions) ToQuery() url.Values {
	q := o.ListOptions.ToQuery()

	if o.UpdatedAfter != "" {
		q.Set("updated_after", o.UpdatedAfter)
	}
	if o.UpdatedBefore != "" {
		q.Set("updated_before", o.UpdatedBefore)
	}
	if o.AuthorID > 0 {
		q.Set("author_id", fmt.Sprintf("%d", o.AuthorID))
	}
	if o.AssigneeID > 0 {
		q.Set("assignee_id", fmt.Sprintf("%d", o.AssigneeID))
	}
	if o.Labels != "" {
		q.Set("labels", o.Labels)
	}
	if o.Scope != "" {
		q.Set("scope", o.Scope)
	}

	return q
}

func (o *PipelineListOptions) ToQuery() url.Values {
	q := o.ListOptions.ToQuery()

	if o.UpdatedAfter != "" {
		q.Set("updated_after", o.UpdatedAfter)
	}
	if o.UpdatedBefore != "" {
		q.Set("updated_before", o.UpdatedBefore)
	}
	if o.Status != "" {
		q.Set("status", o.Status)
	}
	if o.Ref != "" {
		q.Set("ref", o.Ref)
	}
	if o.Source != "" {
		q.Set("source", o.Source)
	}

	return q
}

// SinceHours returns an ISO8601 timestamp for N hours ago.
// Useful for time-based filtering.
func SinceHours(hours int) string {
	t := time.Now().Add(-time.Duration(hours) * time.Hour)
	return t.UTC().Format(time.RFC3339)
}

// SinceDays returns an ISO8601 timestamp for N days ago.
func SinceDays(days int) string {
	return SinceHours(days * 24)
}
