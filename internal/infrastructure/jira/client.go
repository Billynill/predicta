package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/predicta/predicta/internal/domain/entity"
)

type Client struct {
	baseURL    string
	email      string
	apiToken   string
	sprintID   string
	projectKey string
	httpClient *http.Client
}

func NewClient(baseURL, email, apiToken, sprintID, projectKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		email:      email,
		apiToken:   apiToken,
		sprintID:   sprintID,
		projectKey: projectKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) GetCurrentSprint(ctx context.Context, _ string) (*entity.Sprint, error) {
	// TODO: GET /rest/agile/1.0/sprint/{sprintId}
	return &entity.Sprint{
		ID:            c.sprintID,
		Name:          "Sprint from Jira",
		DaysRemaining: 3,
	}, nil
}

func (c *Client) GetSprintTasks(ctx context.Context, sprintID string) ([]entity.Task, error) {
	url := fmt.Sprintf("%s/rest/agile/1.0/sprint/%s/issue", c.baseURL, sprintID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.email, c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jira sprint issues: status %d", resp.StatusCode)
	}

	// TODO: map Jira issues -> entity.Task
	return nil, fmt.Errorf("jira issue mapping not implemented")
}

func (c *Client) ReassignTask(ctx context.Context, taskExternalID, newAssigneeExternalID string) error {
	payload := map[string]any{
		"fields": map[string]any{
			"assignee": map[string]string{"accountId": newAssigneeExternalID},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/rest/api/3/issue/%s", c.baseURL, taskExternalID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.email, c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("jira reassign: status %d", resp.StatusCode)
	}
	return nil
}
