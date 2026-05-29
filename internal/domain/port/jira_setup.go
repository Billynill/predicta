package port

import (
	"context"

	"github.com/predicta/predicta/internal/domain/entity"
)

// JiraSetupUser — пользователь Jira для настройки интеграции.
type JiraSetupUser struct {
	AccountID   string `json:"account_id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// JiraSetupProject — проект Jira.
type JiraSetupProject struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// JiraSetupSprint — метаданные спринта.
type JiraSetupSprint struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	State      string `json:"state"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	ProjectKey string `json:"project_key"`
}

// JiraSetup — dev/setup ручки Jira (не core domain flow).
type JiraSetup interface {
	WhoAmI(ctx context.Context) (*JiraSetupUser, error)
	ListProjects(ctx context.Context) ([]JiraSetupProject, error)
	ResolveProjectKey(ctx context.Context) (string, error)
	ListAssignableUsers(ctx context.Context, projectKey string) ([]JiraSetupUser, error)
	GetSprintInfo(ctx context.Context) (*JiraSetupSprint, error)
	GetSprintTasks(ctx context.Context, sprintID string) ([]entity.Task, error)
}
