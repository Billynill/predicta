package usecase

import (
	"context"

	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
)

// ProjectStatusGetter — экран 1: здоровье проекта.
type ProjectStatusGetter interface {
	GetStatus(ctx context.Context) (entity.ProjectStatus, error)
}

type ProjectService struct {
	tracker  port.TaskTracker
	velocity port.VelocityCalculator
	teamID   string
	track    string
}

func NewProjectService(
	tracker port.TaskTracker,
	velocity port.VelocityCalculator,
	teamID, track string,
) *ProjectService {
	return &ProjectService{
		tracker:  tracker,
		velocity: velocity,
		teamID:   resolveTeamID(teamID),
		track:    resolveTrack(track),
	}
}

func (s *ProjectService) GetStatus(ctx context.Context) (entity.ProjectStatus, error) {
	sc, err := loadSprintContext(ctx, s.tracker, s.teamID)
	if err != nil {
		return entity.ProjectStatus{}, err
	}
	return s.velocity.BuildProjectStatus(*sc.sprint, sc.tasks, s.track), nil
}
