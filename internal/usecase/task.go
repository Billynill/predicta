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

// TaskCreator — создание задачи с проверкой загрузки через AI.
type TaskCreator interface {
	Create(ctx context.Context, cmd CreateTaskCommand) (entity.CreateTaskResult, error)
}

type CreateTaskCommand struct {
	Title       string
	Description string
	AssigneeID  string
	Force       bool
}

type TaskService struct {
	tracker   port.TaskTracker
	employees port.EmployeeRepository
	chats     port.ChatRepository
	ai        port.AIAnalyzer
	velocity  port.VelocityCalculator
	teamID    string
	track     string
}

func NewTaskService(
	tracker port.TaskTracker,
	employees port.EmployeeRepository,
	chats port.ChatRepository,
	ai port.AIAnalyzer,
	velocity port.VelocityCalculator,
	teamID, track string,
) *TaskService {
	return &TaskService{
		tracker:   tracker,
		employees: employees,
		chats:     chats,
		ai:        ai,
		velocity:  velocity,
		teamID:    resolveTeamID(teamID),
		track:     resolveTrack(track),
	}
}

func (s *TaskService) Create(ctx context.Context, cmd CreateTaskCommand) (entity.CreateTaskResult, error) {
	assignee, err := s.employees.GetByID(ctx, cmd.AssigneeID)
	if err != nil {
		return entity.CreateTaskResult{}, fmt.Errorf("get assignee: %w", err)
	}

	sc, err := loadSprintContext(ctx, s.tracker, s.teamID)
	if err != nil {
		return entity.CreateTaskResult{}, err
	}

	employees, err := s.employees.ListByTeam(ctx, s.teamID)
	if err != nil {
		return entity.CreateTaskResult{}, err
	}

	velocities := s.velocity.BuildTeamVelocity(employees, sc.tasks)
	decision := evaluateAssigneeLoad(*assignee, velocities)

	members, err := buildTeamMemberInputs(ctx, employees, sc.tasks, s.chats)
	if err != nil {
		return entity.CreateTaskResult{}, err
	}

	requestedMember := teamMemberForEmployee(*assignee, members)
	suggestedMember := teamMemberForEmployee(decision.Suggested, members)

	aiInsight, err := s.ai.RecommendTaskAssignee(ctx, port.TaskAssigneeRecommendationInput{
		TaskTitle:         cmd.Title,
		TaskDescription:   cmd.Description,
		SprintName:        sc.sprint.Name,
		DaysLeft:          sc.sprint.DaysRemaining,
		Approved:          decision.Approved,
		RequestedAssignee: requestedMember,
		SuggestedAssignee: suggestedMember,
		TeamMembers:       members,
	})
	if err != nil {
		return entity.CreateTaskResult{}, fmt.Errorf("ai recommend assignee: %w", err)
	}

	result := entity.CreateTaskResult{
		Approved:              decision.Approved,
		AssigneeID:            assignee.ExternalID,
		AssigneeName:          assignee.Name,
		SuggestedAssigneeID:   decision.Suggested.ExternalID,
		SuggestedAssigneeName: decision.Suggested.Name,
		AIInsight:             aiInsight,
	}

	if !decision.Approved && !cmd.Force {
		return result, nil
	}

	created, err := s.tracker.CreateTask(ctx, port.CreateTaskInput{
		Title:       cmd.Title,
		Description: cmd.Description,
		AssigneeID:  assignee.ExternalID,
		SprintID:    sc.sprint.ID,
	})
	if err != nil {
		return entity.CreateTaskResult{}, fmt.Errorf("create task in tracker: %w", err)
	}

	result.Created = true
	result.TaskID = created.TaskID
	result.TaskTitle = created.Title
	return result, nil
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