package mock

import (
	"context"
	"fmt"

	"github.com/predicta/predicta/internal/domain/entity"
)

const SprintID = "sprint-5"

// Tracker — демо-задачи для хакатона; assignee берутся из config/employees.json.
type Tracker struct {
	teamID              string
	sprintDaysRemaining int
	tasks               map[string]entity.Task
	members             []entity.Employee
}

func NewTracker(teamID string, sprintDaysRemaining int, members []entity.Employee) *Tracker {
	t := &Tracker{
		teamID:              teamID,
		sprintDaysRemaining: sprintDaysRemaining,
		tasks:               make(map[string]entity.Task),
		members:             members,
	}
	t.seed()
	return t
}

func (t *Tracker) seed() {
	if len(t.members) < 2 {
		return
	}
	fast := t.members[0]
	slow := t.members[1]

	seeds := []taskSeed{
		{id: "task-1", title: "Auth API", assignee: fast.ID, status: entity.TaskStatusDone},
		{id: "task-2", title: "Users CRUD", assignee: fast.ID, status: entity.TaskStatusDone},
		{id: "task-3", title: "Notifications", assignee: fast.ID, status: entity.TaskStatusDone},
		{id: "task-4", title: "Metrics export", assignee: fast.ID, status: entity.TaskStatusDone},
		{id: "task-5", title: "Release checklist", assignee: fast.ID, status: entity.TaskStatusDone},
		{id: "task-6", title: "DB migrations", assignee: slow.ID, status: entity.TaskStatusDone},
		{id: "task-7", title: "Payment gateway", assignee: slow.ID, status: entity.TaskStatusInProgress},
		{id: "task-8", title: "Webhook handler", assignee: slow.ID, status: entity.TaskStatusTodo},
		{id: "task-9", title: "Retry policy", assignee: slow.ID, status: entity.TaskStatusTodo},
		{id: "task-10", title: "Load tests", assignee: slow.ID, status: entity.TaskStatusTodo},
	}

	for _, item := range seeds {
		t.tasks[item.id] = entity.Task{
			ID:         item.id,
			ExternalID: item.id,
			Title:      item.title,
			AssigneeID: item.assignee,
			Status:     item.status,
			SprintID:   SprintID,
		}
	}
}

func (t *Tracker) GetCurrentSprint(_ context.Context, _ string) (*entity.Sprint, error) {
	return &entity.Sprint{
		ID:            SprintID,
		Name:          "Спринт №5",
		TeamID:        t.teamID,
		DaysRemaining: t.sprintDaysRemaining,
	}, nil
}

func (t *Tracker) GetSprintTasks(_ context.Context, sprintID string) ([]entity.Task, error) {
	if sprintID != SprintID {
		return nil, fmt.Errorf("sprint not found: %s", sprintID)
	}
	out := make([]entity.Task, 0, len(t.tasks))
	for _, task := range t.tasks {
		out = append(out, task)
	}
	return out, nil
}

func (t *Tracker) ReassignTask(_ context.Context, taskExternalID, newAssigneeExternalID string) error {
	task, ok := t.tasks[taskExternalID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskExternalID)
	}
	task.AssigneeID = resolveAssigneeID(newAssigneeExternalID, t.members)
	t.tasks[taskExternalID] = task
	return nil
}

func resolveAssigneeID(ref string, members []entity.Employee) string {
	for _, m := range members {
		if ref == m.ID || ref == m.ExternalID {
			return m.ID
		}
	}
	return ref
}

type taskSeed struct {
	id, title, assignee string
	status              entity.TaskStatus
}
