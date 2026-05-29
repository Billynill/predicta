package usecase

import (
	"context"

	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
)

// TeamVelocityGetter — экран 2: velocity команды.
type TeamVelocityGetter interface {
	GetVelocity(ctx context.Context) ([]entity.EmployeeVelocity, error)
}

// TeamInsightsGetter — AI-анализ загрузки всей команды.
type TeamInsightsGetter interface {
	GetInsights(ctx context.Context) (string, error)
}

type TeamService struct {
	tracker   port.TaskTracker
	employees port.EmployeeRepository
	chats     port.ChatRepository
	ai        port.AIAnalyzer
	velocity  port.VelocityCalculator
	teamID    string
}

func NewTeamService(
	tracker port.TaskTracker,
	employees port.EmployeeRepository,
	chats port.ChatRepository,
	ai port.AIAnalyzer,
	velocity port.VelocityCalculator,
	teamID string,
) *TeamService {
	return &TeamService{
		tracker:   tracker,
		employees: employees,
		chats:     chats,
		ai:        ai,
		velocity:  velocity,
		teamID:    resolveTeamID(teamID),
	}
}

func (s *TeamService) GetVelocity(ctx context.Context) ([]entity.EmployeeVelocity, error) {
	sc, err := loadSprintContext(ctx, s.tracker, s.teamID)
	if err != nil {
		return nil, err
	}

	employees, err := s.employees.ListByTeam(ctx, s.teamID)
	if err != nil {
		return nil, err
	}

	return s.velocity.BuildTeamVelocity(employees, sc.tasks), nil
}

func (s *TeamService) GetInsights(ctx context.Context) (string, error) {
	sc, err := loadSprintContext(ctx, s.tracker, s.teamID)
	if err != nil {
		return "", err
	}

	employees, err := s.employees.ListByTeam(ctx, s.teamID)
	if err != nil {
		return "", err
	}

	members, err := buildTeamMemberInputs(ctx, employees, sc.tasks, s.chats)
	if err != nil {
		return "", err
	}

	return s.ai.AnalyzeTeamWorkload(ctx, port.TeamAnalysisInput{
		SprintName: sc.sprint.Name,
		DaysLeft:   sc.sprint.DaysRemaining,
		Members:    members,
	})
}
