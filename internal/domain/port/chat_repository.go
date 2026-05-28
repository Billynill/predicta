package port

import (
	"context"

	"github.com/predicta/predicta/internal/domain/entity"
)

// ChatRepository — логи рабочего чата (Telegram).
type ChatRepository interface {
	Save(ctx context.Context, msg entity.ChatMessage) error
	ListRecentByEmployee(ctx context.Context, employeeID string, limit int) ([]entity.ChatMessage, error)
}
