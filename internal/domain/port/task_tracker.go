package port

import (
	"context"

	"github.com/predicta/predicta/internal/domain/entity"
)

// TaskTracker — порт к Jira / Linear.
type TaskTracker interface {
	GetCurrentSprint(ctx context.Context, teamID string) (*entity.Sprint, error)
	GetSprintTasks(ctx context.Context, sprintID string) ([]entity.Task, error)
	ReassignTask(ctx context.Context, taskExternalID, newAssigneeExternalID string) error
}
