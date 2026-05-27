package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
	"github.com/predicta/predicta/internal/domain/service"
)

const (
	defaultTeamID   = "backend"
	defaultTrack    = "бэкенда"
	chatLogLimit    = 10
)

type ProjectService struct {
	tracker  port.TaskTracker
	velocity *service.VelocityEngine
	teamID   string
	track    string
}

func NewProjectService(tracker port.TaskTracker, velocity *service.VelocityEngine, teamID, track string) *ProjectService {
	if teamID == "" {
		teamID = defaultTeamID
	}
	if track == "" {
		track = defaultTrack
	}
	return &ProjectService{
		tracker:  tracker,
		velocity: velocity,
		teamID:   teamID,
		track:    track,
	}
}

func (s *ProjectService) GetStatus(ctx context.Context) (entity.ProjectStatus, error) {
	sprint, err := s.tracker.GetCurrentSprint(ctx, s.teamID)
	if err != nil {
		return entity.ProjectStatus{}, fmt.Errorf("get sprint: %w", err)
	}

	tasks, err := s.tracker.GetSprintTasks(ctx, sprint.ID)
	if err != nil {
		return entity.ProjectStatus{}, fmt.Errorf("get sprint tasks: %w", err)
	}

	return s.velocity.BuildProjectStatus(*sprint, tasks, s.track), nil
}

type TeamService struct {
	tracker   port.TaskTracker
	employees port.EmployeeStore
	velocity  *service.VelocityEngine
	teamID    string
}

func NewTeamService(
	tracker port.TaskTracker,
	employees port.EmployeeStore,
	velocity *service.VelocityEngine,
	teamID string,
) *TeamService {
	if teamID == "" {
		teamID = defaultTeamID
	}
	return &TeamService{
		tracker:   tracker,
		employees: employees,
		velocity:  velocity,
		teamID:    teamID,
	}
}

func (s *TeamService) GetVelocity(ctx context.Context) ([]entity.EmployeeVelocity, error) {
	sprint, err := s.tracker.GetCurrentSprint(ctx, s.teamID)
	if err != nil {
		return nil, fmt.Errorf("get sprint: %w", err)
	}

	employees, err := s.employees.ListByTeam(ctx, s.teamID)
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", err)
	}

	tasks, err := s.tracker.GetSprintTasks(ctx, sprint.ID)
	if err != nil {
		return nil, fmt.Errorf("get sprint tasks: %w", err)
	}

	return s.velocity.BuildTeamVelocity(employees, tasks), nil
}

type EmployeeService struct {
	tracker   port.TaskTracker
	employees port.EmployeeStore
	chats     port.ChatStore
	ai        port.AIAnalyzer
	velocity  *service.VelocityEngine
	teamID    string
}

func NewEmployeeService(
	tracker port.TaskTracker,
	employees port.EmployeeStore,
	chats port.ChatStore,
	ai port.AIAnalyzer,
	velocity *service.VelocityEngine,
	teamID string,
) *EmployeeService {
	if teamID == "" {
		teamID = defaultTeamID
	}
	return &EmployeeService{
		tracker:   tracker,
		employees: employees,
		chats:     chats,
		ai:        ai,
		velocity:  velocity,
		teamID:    teamID,
	}
}

func (s *EmployeeService) GetAnalytics(ctx context.Context, employeeID string) (entity.EmployeeAnalytics, error) {
	employee, err := s.employees.GetByID(ctx, employeeID)
	if err != nil {
		return entity.EmployeeAnalytics{}, fmt.Errorf("get employee: %w", err)
	}

	sprint, err := s.tracker.GetCurrentSprint(ctx, s.teamID)
	if err != nil {
		return entity.EmployeeAnalytics{}, fmt.Errorf("get sprint: %w", err)
	}

	tasks, err := s.tracker.GetSprintTasks(ctx, sprint.ID)
	if err != nil {
		return entity.EmployeeAnalytics{}, fmt.Errorf("get sprint tasks: %w", err)
	}

	employeeTasks := filterByAssignee(tasks, employee.ID)
	forecast := s.velocity.BuildEmployeeForecast(*employee, tasks, sprint.DaysRemaining)

	logs, err := s.chats.ListRecentByEmployee(ctx, employee.ID, chatLogLimit)
	if err != nil {
		return entity.EmployeeAnalytics{}, fmt.Errorf("list chat logs: %w", err)
	}

	done, total := countAssignee(employeeTasks)
	insight, err := s.ai.AnalyzeEmployeeContext(ctx, port.EmployeeAnalysisInput{
		EmployeeName: employee.Name,
		Role:         employee.Role,
		DoneCount:    done,
		TotalCount:   total,
		DelayDays:    forecast.DelayDays,
		ChatLogs:     formatChatLogs(logs),
	})
	if err != nil {
		return entity.EmployeeAnalytics{}, fmt.Errorf("ai analyze: %w", err)
	}

	return entity.EmployeeAnalytics{
		Employee:  *employee,
		Forecast:  forecast,
		Tasks:     employeeTasks,
		ChatLogs:  logs,
		AIInsight: insight,
	}, nil
}

type TaskService struct {
	tracker   port.TaskTracker
	employees port.EmployeeStore
	velocity  *service.VelocityEngine
	project   *ProjectService
	teamID    string
	track     string
}

func NewTaskService(
	tracker port.TaskTracker,
	employees port.EmployeeStore,
	velocity *service.VelocityEngine,
	project *ProjectService,
	teamID, track string,
) *TaskService {
	if teamID == "" {
		teamID = defaultTeamID
	}
	if track == "" {
		track = defaultTrack
	}
	return &TaskService{
		tracker:   tracker,
		employees: employees,
		velocity:  velocity,
		project:   project,
		teamID:    teamID,
		track:     track,
	}
}

func (s *TaskService) Reassign(ctx context.Context, taskID, newExecutorID string) (entity.ReassignResult, error) {
	newEmployee, err := s.employees.GetByID(ctx, newExecutorID)
	if err != nil {
		return entity.ReassignResult{}, fmt.Errorf("get new executor: %w", err)
	}

	sprint, err := s.tracker.GetCurrentSprint(ctx, s.teamID)
	if err != nil {
		return entity.ReassignResult{}, fmt.Errorf("get sprint: %w", err)
	}

	tasks, err := s.tracker.GetSprintTasks(ctx, sprint.ID)
	if err != nil {
		return entity.ReassignResult{}, fmt.Errorf("get sprint tasks: %w", err)
	}

	var task *entity.Task
	for i := range tasks {
		if tasks[i].ID == taskID || tasks[i].ExternalID == taskID {
			task = &tasks[i]
			break
		}
	}
	if task == nil {
		return entity.ReassignResult{}, fmt.Errorf("task not found: %s", taskID)
	}

	if err := s.tracker.ReassignTask(ctx, task.ExternalID, newEmployee.ExternalID); err != nil {
		return entity.ReassignResult{}, fmt.Errorf("reassign in tracker: %w", err)
	}

	updatedTasks, err := s.tracker.GetSprintTasks(ctx, sprint.ID)
	if err != nil {
		return entity.ReassignResult{}, fmt.Errorf("refresh sprint tasks: %w", err)
	}

	status := s.velocity.BuildProjectStatus(*sprint, updatedTasks, s.track)
	msg := fmt.Sprintf(
		"Задача перенаправлена %s. Новый прогноз проекта: %s",
		newEmployee.Name,
		reassignOutcome(status),
	)

	return entity.ReassignResult{
		TaskID:          task.ID,
		TaskTitle:       task.Title,
		NewAssigneeID:   newEmployee.ID,
		NewAssigneeName: newEmployee.Name,
		ProjectStatus:   status,
		Message:         msg,
	}, nil
}

func filterByAssignee(tasks []entity.Task, assigneeID string) []entity.Task {
	out := make([]entity.Task, 0)
	for _, t := range tasks {
		if t.AssigneeID == assigneeID {
			out = append(out, t)
		}
	}
	return out
}

func countAssignee(tasks []entity.Task) (done, total int) {
	total = len(tasks)
	for _, t := range tasks {
		if t.Status == entity.TaskStatusDone {
			done++
		}
	}
	return done, total
}

func formatChatLogs(logs []entity.ChatMessage) []string {
	out := make([]string, 0, len(logs))
	for _, l := range logs {
		out = append(out, fmt.Sprintf("%s: %s", l.SentAt.Format(time.RFC3339), l.Text))
	}
	return out
}

func reassignOutcome(status entity.ProjectStatus) string {
	if status.IsAtRisk {
		return fmt.Sprintf("задержка сократилась до %d дн.", status.DelayDays)
	}
	return "Сдача вовремя"
}
