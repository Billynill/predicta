package port

import (
	"context"

	"github.com/predicta/predicta/internal/domain/entity"
)

// TaskTracker — порт к Jira / Linear.
type TaskTracker interface {
	GetCurrentSprint(ctx context.Context, teamID string) (*entity.Sprint, error)
	GetSprintTasks(ctx context.Context, sprintID string) ([]entity.Task, error)
	ReassignTask(ctx context.Context, taskExternalID, newAssigneeExternalID string) error
}

// EmployeeStore — маппинг сотрудников и telegram-ников.
type EmployeeStore interface {
	ListByTeam(ctx context.Context, teamID string) ([]entity.Employee, error)
	GetByID(ctx context.Context, id string) (*entity.Employee, error)
	GetByTelegramNick(ctx context.Context, nick string) (*entity.Employee, error)
}

// ChatStore — сохранение сообщений из Telegram.
type ChatStore interface {
	Save(ctx context.Context, msg entity.ChatMessage) error
	ListRecentByEmployee(ctx context.Context, employeeID string, limit int) ([]entity.ChatMessage, error)
}

// AIAnalyzer — GigaChat.
type AIAnalyzer interface {
	AnalyzeEmployeeContext(ctx context.Context, input EmployeeAnalysisInput) (string, error)
}

type EmployeeAnalysisInput struct {
	EmployeeName string
	Role         string
	DoneCount    int
	TotalCount   int
	DelayDays    int
	ChatLogs     []string
}
