package usecase

import (
	"context"
	"fmt"

	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
)

// ProjectStatusGetter — экран 1: здоровье проекта.
type ProjectStatusGetter interface {
	GetStatus(ctx context.Context) (entity.ProjectStatus, error)
}

type ProjectService struct {
	tracker   port.TaskTracker
	employees port.EmployeeRepository
	chats     port.ChatRepository
	ai        port.AIAnalyzer
	velocity  port.VelocityCalculator
	teamID    string
	track     string
}

func NewProjectService(
	tracker port.TaskTracker,
	employees port.EmployeeRepository,
	chats port.ChatRepository,
	ai port.AIAnalyzer,
	velocity port.VelocityCalculator,
	teamID, track string,
) *ProjectService {
	return &ProjectService{
		tracker:   tracker,
		employees: employees,
		chats:     chats,
		ai:        ai,
		velocity:  velocity,
		teamID:    resolveTeamID(teamID),
		track:     resolveTrack(track),
	}
}

func (s *ProjectService) GetStatus(ctx context.Context) (entity.ProjectStatus, error) {
	sc, err := loadSprintContext(ctx, s.tracker, s.teamID)
	if err != nil {
		return entity.ProjectStatus{}, err
	}

	status := s.velocity.BuildProjectStatus(*sc.sprint, sc.tasks, s.track)

	employees, err := s.employees.ListByTeam(ctx, s.teamID)
	if err != nil {
		return entity.ProjectStatus{}, err
	}

	members, err := buildTeamMemberInputs(ctx, employees, sc.tasks, s.chats)
	if err != nil {
		return entity.ProjectStatus{}, err
	}

	done, total := countTaskProgress(sc.tasks)
	remaining := total - done

	advice, err := s.ai.AnalyzeProjectStatus(ctx, port.ProjectStatusInput{
		SprintName:     status.SprintName,
		TrackName:      status.TrackName,
		DaysRemaining:  status.DaysRemaining,
		CompletionPct:  status.CompletionPct,
		DelayDays:      status.DelayDays,
		IsAtRisk:       status.IsAtRisk,
		RiskMessage:    status.RiskMessage,
		RemainingTasks: remaining,
		TotalTasks:     total,
		DoneTasks:      done,
		TeamMembers:    members,
	})
	if err != nil {
		return entity.ProjectStatus{}, fmt.Errorf("ai project status: %w", err)
	}

	status.AIAdvice = advice
	return status, nil
}
