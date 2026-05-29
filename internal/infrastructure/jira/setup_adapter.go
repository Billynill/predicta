package jira

import (
	"context"

	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
)

// SetupAdapter — реализация port.JiraSetup поверх Client.
type SetupAdapter struct {
	client *Client
}

func NewSetupAdapter(client *Client) *SetupAdapter {
	return &SetupAdapter{client: client}
}

var (
	_ port.JiraSetup   = (*SetupAdapter)(nil)
	_ port.TaskTracker = (*Client)(nil)
	_ port.UserLookup  = (*Client)(nil)
)

func (a *SetupAdapter) WhoAmI(ctx context.Context) (*port.JiraSetupUser, error) {
	u, err := a.client.WhoAmI(ctx)
	if err != nil {
		return nil, err
	}
	return mapSetupUser(u), nil
}

func (a *SetupAdapter) ListProjects(ctx context.Context) ([]port.JiraSetupProject, error) {
	projects, err := a.client.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]port.JiraSetupProject, 0, len(projects))
	for _, p := range projects {
		out = append(out, port.JiraSetupProject{
			ID:   p.ID,
			Key:  p.Key,
			Name: p.Name,
		})
	}
	return out, nil
}

func (a *SetupAdapter) ResolveProjectKey(ctx context.Context) (string, error) {
	return a.client.ResolveProjectKey(ctx)
}

func (a *SetupAdapter) ListAssignableUsers(ctx context.Context, projectKey string) ([]port.JiraSetupUser, error) {
	users, err := a.client.ListAssignableUsers(ctx, projectKey)
	if err != nil {
		return nil, err
	}
	out := make([]port.JiraSetupUser, 0, len(users))
	for _, u := range users {
		out = append(out, *mapSetupUser(&u))
	}
	return out, nil
}

func (a *SetupAdapter) GetSprintInfo(ctx context.Context) (*port.JiraSetupSprint, error) {
	sprint, err := a.client.GetSprintInfo(ctx)
	if err != nil {
		return nil, err
	}
	projectKey, _ := a.client.ResolveProjectKey(ctx)
	return &port.JiraSetupSprint{
		ID:         sprint.ID,
		Name:       sprint.Name,
		State:      sprint.State,
		StartDate:  sprint.StartDate,
		EndDate:    sprint.EndDate,
		ProjectKey: projectKey,
	}, nil
}

func (a *SetupAdapter) GetSprintTasks(ctx context.Context, sprintID string) ([]entity.Task, error) {
	return a.client.GetSprintTasks(ctx, sprintID)
}

func mapSetupUser(u *AssignableUser) *port.JiraSetupUser {
	if u == nil {
		return nil
	}
	return &port.JiraSetupUser{
		AccountID:   u.AccountID,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		AvatarURL:   u.AvatarURL,
	}
}
