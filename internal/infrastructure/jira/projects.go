package jira

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

// ListProjects — доступные проекты (для выбора JIRA_PROJECT_KEY).
func (c *Client) ListProjects(ctx context.Context) ([]ProjectSummary, error) {
	q := url.Values{}
	q.Set("maxResults", "50")
	q.Set("orderBy", "key")

	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/api/3/project/search", q, nil)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Values []struct {
			ID   string `json:"id"`
			Key  string `json:"key"`
			Name string `json:"name"`
		} `json:"values"`
	}
	if err := c.decodeJSON(resp, &raw); err != nil {
		return nil, err
	}

	out := make([]ProjectSummary, 0, len(raw.Values))
	for _, p := range raw.Values {
		if p.Key == "" {
			continue
		}
		out = append(out, ProjectSummary{ID: p.ID, Key: p.Key, Name: p.Name})
	}
	return out, nil
}

// ResolveProjectKey — ключ проекта из env или из задач текущего спринта.
func (c *Client) ResolveProjectKey(ctx context.Context) (string, error) {
	if key := strings.TrimSpace(c.projectKey); key != "" {
		return key, nil
	}
	return c.resolveProjectKeyFromSprint(ctx)
}

func (c *Client) resolveProjectKeyFromSprint(ctx context.Context) (string, error) {
	tasks, err := c.GetSprintTasks(ctx, "")
	if err != nil {
		return "", err
	}
	for _, t := range tasks {
		if key := projectKeyFromIssueKey(t.ID); key != "" {
			return key, nil
		}
	}
	return "", errNoProjectKey
}

func (c *Client) candidateProjectKeys(ctx context.Context, explicit string) []string {
	seen := map[string]bool{}
	var keys []string
	add := func(k string) {
		k = strings.TrimSpace(k)
		if k == "" || seen[k] {
			return
		}
		seen[k] = true
		keys = append(keys, k)
	}

	add(explicit)
	if explicit == "" {
		add(c.projectKey)
	}
	if derived, err := c.resolveProjectKeyFromSprint(ctx); err == nil {
		add(derived)
	}
	return keys
}

func (c *Client) listAssignableUsersForProject(ctx context.Context, projectKey string) ([]AssignableUser, error) {
	q := url.Values{}
	q.Set("project", projectKey)
	q.Set("maxResults", "100")

	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/api/3/user/assignable/search", q, nil)
	if err != nil {
		return nil, err
	}

	var raw []jiraUser
	if err := c.decodeJSON(resp, &raw); err != nil {
		return nil, err
	}

	out := make([]AssignableUser, 0, len(raw))
	for _, u := range raw {
		if u.AccountID == "" {
			continue
		}
		out = append(out, mapAssignable(u))
	}
	return out, nil
}

func projectKeyFromIssueKey(issueKey string) string {
	if i := strings.IndexByte(issueKey, '-'); i > 0 {
		return issueKey[:i]
	}
	return ""
}

var errNoProjectKey = errString("jira: could not detect project key from sprint tasks")

type errString string

func (e errString) Error() string { return string(e) }

func isUnknownProjectErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "No project with the provided key") ||
		strings.Contains(msg, "No project could be found")
}
