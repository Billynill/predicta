package app

import (
	"github.com/predicta/predicta/config"
	"github.com/predicta/predicta/internal/domain/port"
	"github.com/predicta/predicta/internal/infrastructure/employees"
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
	JiraSetup port.JiraSetup
	AI        port.AIAnalyzer
	Velocity  port.VelocityCalculator
	Tokens    port.TokenIssuer
	Managers  port.ManagerRepository

	Auth    *usecase.AuthService
	Project usecase.ProjectStatusGetter
	Team    usecase.TeamVelocityGetter
	TeamAI  usecase.TeamInsightsGetter
	Employee usecase.EmployeeAnalyticsGetter
	Profile  usecase.EmployeeProfileGetter
	Task     usecase.TaskReassigner
	Creator  usecase.TaskCreator
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
