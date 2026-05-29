package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/predicta/predicta/internal/domain/entity"
)

type Client struct {
	baseURL            string
	email              string
	apiToken           string
	sprintID           string
	projectKey         string
	sprintDaysFallback int
	httpClient         *http.Client
}

func NewClient(
	baseURL, email, apiToken, sprintID, projectKey string,
	sprintDaysFallback int,
) *Client {
	if sprintDaysFallback <= 0 {
		sprintDaysFallback = 3
	}
	return &Client{
		baseURL:            normalizeBaseURL(baseURL),
		email:              email,
		apiToken:           apiToken,
		sprintID:           sprintID,
		projectKey:         projectKey,
		sprintDaysFallback: sprintDaysFallback,
		httpClient:         &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) GetCurrentSprint(ctx context.Context, _ string) (*entity.Sprint, error) {
	sprint, err := c.resolveSprint(ctx)
	if err != nil {
		return nil, err
	}

	tasks, err := c.GetSprintTasks(ctx, sprintIDString(sprint.ID))
	if err != nil {
		return nil, err
	}

	done := 0
	for _, t := range tasks {
		if t.Status == entity.TaskStatusDone {
			done++
		}
	}

	return &entity.Sprint{
		ID:             sprintIDString(sprint.ID),
		Name:           fmtSprintName(*sprint),
		DaysRemaining:  sprintDaysRemaining(sprint.EndDate, c.sprintDaysFallback),
		TotalTasks:     len(tasks),
		CompletedTasks: done,
	}, nil
}

func (c *Client) GetSprintTasks(ctx context.Context, sprintID string) ([]entity.Task, error) {
	if sprintID == "" {
		sprint, err := c.resolveSprint(ctx)
		if err != nil {
			return nil, err
		}
		sprintID = sprintIDString(sprint.ID)
	}

	var all []entity.Task
	startAt := 0
	const pageSize = 50

	for {
		q := url.Values{}
		q.Set("startAt", strconv.Itoa(startAt))
		q.Set("maxResults", strconv.Itoa(pageSize))
		q.Set("fields", sprintIssueFields)

		path := fmt.Sprintf("/rest/agile/1.0/sprint/%s/issue", url.PathEscape(sprintID))
		resp, err := c.doRequest(ctx, http.MethodGet, path, q, nil)
		if err != nil {
			return nil, err
		}

		var page issuesPageResponse
		if err := c.decodeIssuesPage(resp, &page); err != nil {
			return nil, err
		}

		for _, issue := range page.Issues {
			all = append(all, mapIssueToTask(issue, sprintID))
		}

		if startAt+page.MaxResults >= page.Total || len(page.Issues) == 0 {
			break
		}
		startAt += page.MaxResults
	}

	return all, nil
}

func (c *Client) ReassignTask(ctx context.Context, taskExternalID, newAssigneeExternalID string) error {
	payload := map[string]any{
		"fields": map[string]any{
			"assignee": map[string]string{"accountId": newAssigneeExternalID},
		},
	}
	path := fmt.Sprintf("/rest/api/3/issue/%s", url.PathEscape(taskExternalID))
	resp, err := c.doRequest(ctx, http.MethodPut, path, nil, payload)
	if err != nil {
		return err
	}
	return c.decodeJSON(resp, nil)
}

// ListAssignableUsers — assignable users в проекте (для employees.json).
func (c *Client) ListAssignableUsers(ctx context.Context, projectKey string) ([]AssignableUser, error) {
	keys := c.candidateProjectKeys(ctx, projectKey)
	if len(keys) == 0 {
		return nil, errNoProjectKey
	}

	var lastErr error
	for _, key := range keys {
		users, err := c.listAssignableUsersForProject(ctx, key)
		if err == nil {
			return users, nil
		}
		lastErr = err
		if !isUnknownProjectErr(err) {
			return nil, err
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errNoProjectKey
}

// GetSprintInfo — метаданные спринта для отладки/настройки.
func (c *Client) GetSprintInfo(ctx context.Context) (*sprintResponse, error) {
	return c.resolveSprint(ctx)
}

// WhoAmI — проверка учётных данных (email + API token).
func (c *Client) WhoAmI(ctx context.Context) (*AssignableUser, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/api/3/myself", nil, nil)
	if err != nil {
		return nil, err
	}
	var u jiraUser
	if err := c.decodeJSON(resp, &u); err != nil {
		return nil, err
	}
	out := mapAssignable(u)
	return &out, nil
}

func (c *Client) resolveSprint(ctx context.Context) (*sprintResponse, error) {
	if c.sprintID != "" {
		return c.fetchSprint(ctx, c.sprintID)
	}
	if c.projectKey == "" {
		return nil, fmt.Errorf("jira: set JIRA_SPRINT_ID or JIRA_PROJECT_KEY")
	}
	return c.findActiveSprint(ctx, c.projectKey)
}

func (c *Client) fetchSprint(ctx context.Context, sprintID string) (*sprintResponse, error) {
	path := fmt.Sprintf("/rest/agile/1.0/sprint/%s", url.PathEscape(sprintID))
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	var sprint sprintResponse
	if err := c.decodeJSON(resp, &sprint); err != nil {
		return nil, err
	}
	return &sprint, nil
}

func (c *Client) findActiveSprint(ctx context.Context, projectKey string) (*sprintResponse, error) {
	q := url.Values{}
	q.Set("projectKeyOrId", projectKey)
	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/agile/1.0/board", q, nil)
	if err != nil {
		return nil, err
	}
	var boards boardListResponse
	if err := c.decodeJSON(resp, &boards); err != nil {
		return nil, err
	}
	if len(boards.Values) == 0 {
		return nil, fmt.Errorf("jira: no board for project %q", projectKey)
	}

	boardID := boards.Values[0].ID
	sprintQ := url.Values{}
	sprintQ.Set("state", "active")
	path := fmt.Sprintf("/rest/agile/1.0/board/%d/sprint", boardID)
	resp, err = c.doRequest(ctx, http.MethodGet, path, sprintQ, nil)
	if err != nil {
		return nil, err
	}
	var list sprintListResponse
	if err := c.decodeJSON(resp, &list); err != nil {
		return nil, err
	}
	if len(list.Values) == 0 {
		return nil, fmt.Errorf("jira: no active sprint on board %d", boardID)
	}
	sprint := list.Values[0]
	return &sprint, nil
}

func (c *Client) decodeIssuesPage(resp *http.Response, page *issuesPageResponse) error {
	defer resp.Body.Close()
	data, err := ioReadAll(resp)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("jira sprint issues: %s", string(data))
	}

	var raw struct {
		StartAt    int `json:"startAt"`
		MaxResults int `json:"maxResults"`
		Total      int `json:"total"`
		Issues     []struct {
			ID     string `json:"id"`
			Key    string `json:"key"`
			Fields map[string]json.RawMessage `json:"fields"`
		} `json:"issues"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	page.StartAt = raw.StartAt
	page.MaxResults = raw.MaxResults
	page.Total = raw.Total
	page.Issues = make([]issueRecord, 0, len(raw.Issues))

	for _, item := range raw.Issues {
		rec := issueRecord{ID: item.ID, Key: item.Key}
		var fields map[string]any
		if err := json.Unmarshal(mustMarshal(item.Fields), &fields); err == nil {
			rec.Fields.Points = extractStoryPoints(fields)
		}
		if summary, ok := item.Fields["summary"]; ok {
			_ = json.Unmarshal(summary, &rec.Fields.Summary)
		}
		if status, ok := item.Fields["status"]; ok {
			_ = json.Unmarshal(status, &rec.Fields.Status)
		}
		if assignee, ok := item.Fields["assignee"]; ok && string(assignee) != "null" {
			var u jiraUser
			if err := json.Unmarshal(assignee, &u); err == nil {
				rec.Fields.Assignee = &u
			}
		}
		page.Issues = append(page.Issues, rec)
	}
	return nil
}

func mustMarshal(m map[string]json.RawMessage) []byte {
	out := make(map[string]any, len(m))
	for k, v := range m {
		var val any
		_ = json.Unmarshal(v, &val)
		out[k] = val
	}
	b, _ := json.Marshal(out)
	return b
}

func ioReadAll(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	const maxBody = 8 << 20
	return io.ReadAll(io.LimitReader(resp.Body, maxBody))
}
