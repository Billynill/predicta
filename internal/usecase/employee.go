package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
)

// EmployeeAnalyticsGetter — экран 3: карточка сотрудника + AI.
type EmployeeAnalyticsGetter interface {
	GetAnalytics(ctx context.Context, employeeID string) (entity.EmployeeAnalytics, error)
}

// EmployeeProfileGetter — профиль сотрудника по account_id + AI-вердикт.
type EmployeeProfileGetter interface {
	GetProfile(ctx context.Context, employeeID string) (entity.EmployeeProfile, error)
}

type EmployeeService struct {
	tracker    port.TaskTracker
	employees  port.EmployeeRepository
	chats      port.ChatRepository
	users      port.UserLookup
	ai         port.AIAnalyzer
	velocity   port.VelocityCalculator
	teamID     string
}

func NewEmployeeService(
	tracker port.TaskTracker,
	employees port.EmployeeRepository,
	chats port.ChatRepository,
	users port.UserLookup,
	ai port.AIAnalyzer,
	velocity port.VelocityCalculator,
	teamID string,
) *EmployeeService {
	return &EmployeeService{
		tracker:   tracker,
		employees: employees,
		chats:     chats,
		users:     users,
		ai:        ai,
		velocity:  velocity,
		teamID:    resolveTeamID(teamID),
	}
}

func (s *EmployeeService) GetAnalytics(ctx context.Context, employeeID string) (entity.EmployeeAnalytics, error) {
	employee, err := s.employees.GetByID(ctx, employeeID)
	if err != nil {
		return entity.EmployeeAnalytics{}, fmt.Errorf("get employee: %w", err)
	}

	sc, err := loadSprintContext(ctx, s.tracker, s.teamID)
	if err != nil {
		return entity.EmployeeAnalytics{}, err
	}

	employeeTasks := filterTasksByAssignee(sc.tasks, *employee)
	forecast := s.velocity.BuildEmployeeForecast(*employee, sc.tasks, sc.sprint.DaysRemaining)

	logs, err := s.chats.ListRecentByEmployee(ctx, employee.ID, chatLogLimit)
	if err != nil {
		return entity.EmployeeAnalytics{}, fmt.Errorf("list chat logs: %w", err)
	}

	done, total := countTaskProgress(employeeTasks)
	insight, err := s.ai.AnalyzeEmployeeContext(ctx, port.EmployeeAnalysisInput{
		EmployeeName: employee.Name,
		Role:         employee.Role,
		DoneCount:    done,
		TotalCount:   total,
		DelayDays:    forecast.DelayDays,
		Tasks:        mapTasksToSummary(employeeTasks),
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

func (s *EmployeeService) GetProfile(ctx context.Context, employeeID string) (entity.EmployeeProfile, error) {
	employee, err := s.employees.GetByID(ctx, employeeID)
	if err != nil {
		return entity.EmployeeProfile{}, fmt.Errorf("get employee: %w", err)
	}

	sc, err := loadSprintContext(ctx, s.tracker, s.teamID)
	if err != nil {
		return entity.EmployeeProfile{}, err
	}

	employeeTasks := filterTasksByAssignee(sc.tasks, *employee)
	done, total := countTaskProgress(employeeTasks)
	remaining := total - done

	team, err := s.employees.ListByTeam(ctx, s.teamID)
	if err != nil {
		return entity.EmployeeProfile{}, err
	}
	velocities := s.velocity.BuildTeamVelocity(team, sc.tasks)
	health := entity.VelocityHealthGood
	for _, v := range velocities {
		if v.Employee.ID == employee.ID {
			health = v.Health
			break
		}
	}

	logs, err := s.chats.ListRecentByEmployee(ctx, employee.ID, chatLogLimit)
	if err != nil {
		return entity.EmployeeProfile{}, fmt.Errorf("list chat logs: %w", err)
	}

	insight, err := s.ai.AnalyzeEmployeeProfile(ctx, port.EmployeeProfileInput{
		EmployeeName: employee.Name,
		Role:         employee.Role,
		Health:       string(health),
		DoneCount:    done,
		TotalCount:   total,
		Remaining:    remaining,
		Tasks:        mapTasksToSummary(employeeTasks),
		ChatLogs:     formatChatLogs(logs),
	})
	if err != nil {
		return entity.EmployeeProfile{}, fmt.Errorf("ai profile: %w", err)
	}

	avatarURL := lookupAvatarURL(ctx, s.users, employee.ExternalID)

	return entity.EmployeeProfile{
		Employee:   *employee,
		DoneCount:  done,
		TotalCount: total,
		Remaining:  remaining,
		Health:     health,
		AvatarURL:  avatarURL,
		Tasks:      employeeTasks,
		AIInsight:  insight,
	}, nil
}

func lookupAvatarURL(ctx context.Context, users port.UserLookup, accountID string) string {
	if users == nil {
		return ""
	}
	url, err := users.GetAvatarURL(ctx, accountID)
	if err != nil {
		return ""
	}
	return url
}

func filterTasksByAssignee(tasks []entity.Task, emp entity.Employee) []entity.Task {
	out := make([]entity.Task, 0)
	for _, t := range tasks {
		if entity.AssigneeMatches(t.AssigneeID, emp) {
			out = append(out, t)
		}
	}
	return out
}

func countTaskProgress(tasks []entity.Task) (done, total int) {
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

func mapTasksToSummary(tasks []entity.Task) []port.TaskSummary {
	out := make([]port.TaskSummary, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, port.TaskSummary{
			ID:     t.ID,
			Title:  t.Title,
			Status: string(t.Status),
		})
	}
	return out
}
