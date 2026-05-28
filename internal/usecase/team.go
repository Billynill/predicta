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

type TeamService struct {
	tracker    port.TaskTracker
	employees  port.EmployeeRepository
	velocity   port.VelocityCalculator
	teamID     string
}

func NewTeamService(
	tracker port.TaskTracker,
	employees port.EmployeeRepository,
	velocity port.VelocityCalculator,
	teamID string,
) *TeamService {
	return &TeamService{
		tracker:   tracker,
		employees: employees,
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
