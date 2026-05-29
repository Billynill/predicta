package jira

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/predicta/predicta/internal/domain/port"
)

func (c *Client) CreateTask(ctx context.Context, in port.CreateTaskInput) (*port.CreateTaskOutput, error) {
	projectKey, err := c.ResolveProjectKey(ctx)
	if err != nil {
		return nil, err
	}

	sprintID := in.SprintID
	if sprintID == "" {
		sprint, err := c.resolveSprint(ctx)
		if err != nil {
			return nil, err
		}
		sprintID = sprintIDString(sprint.ID)
	}

	fields := map[string]any{
		"project":   map[string]string{"key": projectKey},
		"summary":   in.Title,
		"issuetype": map[string]string{"name": "Задание"},
	}
	if in.Description != "" {
		fields["description"] = textADF(in.Description)
	}
	if in.AssigneeID != "" {
		fields["assignee"] = map[string]string{"accountId": in.AssigneeID}
	}

	createResp, err := c.doRequest(ctx, http.MethodPost, "/rest/api/3/issue", nil, map[string]any{"fields": fields})
	if err != nil {
		return nil, err
	}

	var created struct {
		Key string `json:"key"`
	}
	if err := c.decodeJSON(createResp, &created); err != nil {
		return nil, err
	}
	if created.Key == "" {
		return nil, fmt.Errorf("jira: create issue returned empty key")
	}

	addPath := fmt.Sprintf("/rest/agile/1.0/sprint/%s/issue", url.PathEscape(sprintID))
	addResp, err := c.doRequest(ctx, http.MethodPost, addPath, nil, map[string]any{
		"issues": []string{created.Key},
	})
	if err != nil {
		return nil, fmt.Errorf("add issue to sprint: %w", err)
	}
	if err := c.decodeJSON(addResp, nil); err != nil {
		return nil, fmt.Errorf("add issue to sprint: %w", err)
	}

	return &port.CreateTaskOutput{
		TaskID: created.Key,
		Title:  in.Title,
	}, nil
}

func textADF(text string) map[string]any {
	return map[string]any{
		"type":    "doc",
		"version": 1,
		"content": []map[string]any{
			{
				"type": "paragraph",
				"content": []map[string]any{
					{"type": "text", "text": text},
				},
			},
		},
	}
}
