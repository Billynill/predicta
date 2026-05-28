package app

import (
	"github.com/predicta/predicta/config"
	"github.com/predicta/predicta/internal/infrastructure/employees"
	"github.com/predicta/predicta/internal/domain/port"
	"github.com/predicta/predicta/internal/usecase"
)

// Deps — composition root: все зависимости приложения.
type Deps struct {
	Config   *config.Config
	TeamID   string
	Track    string

	Employees port.EmployeeRepository
	Chats     port.ChatRepository
	Tracker   port.TaskTracker
	AI        port.AIAnalyzer
	Velocity  port.VelocityCalculator

	Project  usecase.ProjectStatusGetter
	Team     usecase.TeamVelocityGetter
	Employee usecase.EmployeeAnalyticsGetter
	Task     usecase.TaskReassigner
}

type rosterLoader interface {
	TeamID() string
}

func teamIDFrom(roster rosterLoader, cfg *config.Config) string {
	if id := roster.TeamID(); id != "" {
		return id
	}
	return cfg.TeamID
}

func loadRoster(cfg *config.Config) (*employees.FileRepository, error) {
	return employees.LoadFile(cfg.EmployeesFile)
}
