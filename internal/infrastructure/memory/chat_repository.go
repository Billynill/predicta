package memory

import (
	"context"
	"sync"

	"github.com/predicta/predicta/internal/domain/entity"
)

// ChatRepository — in-memory хранилище сообщений Telegram.
type ChatRepository struct {
	mu       sync.Mutex
	messages []entity.ChatMessage
}

func NewChatRepository(seed []entity.ChatMessage) *ChatRepository {
	if seed == nil {
		seed = []entity.ChatMessage{}
	}
	return &ChatRepository{messages: seed}
}

func (r *ChatRepository) Save(_ context.Context, msg entity.ChatMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = append(r.messages, msg)
	return nil
}

func (r *ChatRepository) ListRecentByEmployee(_ context.Context, employeeID string, limit int) ([]entity.ChatMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]entity.ChatMessage, 0)
	for i := len(r.messages) - 1; i >= 0 && len(out) < limit; i-- {
		if r.messages[i].EmployeeID == employeeID {
			out = append(out, r.messages[i])
		}
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}
