package usecase

import (
	"context"
	"fmt"

	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
)

// TaskReassigner — экран 4: переназначение задачи.
type TaskReassigner interface {
	Reassign(ctx context.Context, taskID, newExecutorID string) (entity.ReassignResult, error)
}

type TaskService struct {
	tracker    port.TaskTracker
	employees  port.EmployeeRepository
	velocity   port.VelocityCalculator
	teamID     string
	track      string
}

func NewTaskService(
	tracker port.TaskTracker,
	employees port.EmployeeRepository,
	velocity port.VelocityCalculator,
	teamID, track string,
) *TaskService {
	return &TaskService{
		tracker:   tracker,
		employees: employees,
		velocity:  velocity,
		teamID:    resolveTeamID(teamID),
		track:     resolveTrack(track),
	}
}

func (s *TaskService) Reassign(ctx context.Context, taskID, newExecutorID string) (entity.ReassignResult, error) {
	newEmployee, err := s.employees.GetByID(ctx, newExecutorID)
	if err != nil {
		return entity.ReassignResult{}, fmt.Errorf("get new executor: %w", err)
	}

	sc, err := loadSprintContext(ctx, s.tracker, s.teamID)
	if err != nil {
		return entity.ReassignResult{}, err
	}

	task, err := findTask(sc.tasks, taskID)
	if err != nil {
		return entity.ReassignResult{}, err
	}

	if err := s.tracker.ReassignTask(ctx, task.ExternalID, newEmployee.ExternalID); err != nil {
		return entity.ReassignResult{}, fmt.Errorf("reassign in tracker: %w", err)
	}

	updatedTasks, err := s.tracker.GetSprintTasks(ctx, sc.sprint.ID)
	if err != nil {
		return entity.ReassignResult{}, fmt.Errorf("refresh sprint tasks: %w", err)
	}

	status := s.velocity.BuildProjectStatus(*sc.sprint, updatedTasks, s.track)
	return entity.ReassignResult{
		TaskID:          task.ID,
		TaskTitle:       task.Title,
		NewAssigneeID:   newEmployee.ID,
		NewAssigneeName: newEmployee.Name,
		ProjectStatus:   status,
		Message: fmt.Sprintf(
			"Задача перенаправлена %s. Новый прогноз проекта: %s",
			newEmployee.Name,
			reassignOutcome(status),
		),
	}, nil
}

func findTask(tasks []entity.Task, taskID string) (*entity.Task, error) {
	for i := range tasks {
		if tasks[i].ID == taskID || tasks[i].ExternalID == taskID {
			return &tasks[i], nil
		}
	}
	return nil, fmt.Errorf("task not found: %s", taskID)
}

func reassignOutcome(status entity.ProjectStatus) string {
	if status.IsAtRisk {
		return fmt.Sprintf("задержка сократилась до %d дн.", status.DelayDays)
	}
	return "Сдача вовремя"
}
