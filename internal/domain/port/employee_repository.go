package port

import (
	"context"

	"github.com/predicta/predicta/internal/domain/entity"
)

// EmployeeRepository — roster команды и маппинг Telegram → сотрудник.
type EmployeeRepository interface {
	ListByTeam(ctx context.Context, teamID string) ([]entity.Employee, error)
	GetByID(ctx context.Context, id string) (*entity.Employee, error)
	GetByTelegramNick(ctx context.Context, nick string) (*entity.Employee, error)
}
