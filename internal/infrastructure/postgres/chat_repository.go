package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/predicta/predicta/internal/domain/entity"
	_ "github.com/lib/pq"
)

// ChatRepository — PostgreSQL-хранилище сообщений Telegram.
type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(dsn string) (*ChatRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return &ChatRepository{db: db}, nil
}

func (r *ChatRepository) Close() error {
	return r.db.Close()
}

func (r *ChatRepository) Save(ctx context.Context, msg entity.ChatMessage) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO chat_messages (employee_id, message_text, sent_at)
		VALUES ($1, $2, $3)`,
		msg.EmployeeID, msg.Text, msg.SentAt,
	)
	return err
}

func (r *ChatRepository) ListRecentByEmployee(ctx context.Context, employeeID string, limit int) ([]entity.ChatMessage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, message_text, sent_at
		FROM chat_messages
		WHERE employee_id = $1
		ORDER BY sent_at DESC
		LIMIT $2`, employeeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.ChatMessage, 0)
	for rows.Next() {
		var m entity.ChatMessage
		var sentAt time.Time
		if err := rows.Scan(&m.ID, &m.EmployeeID, &m.Text, &sentAt); err != nil {
			return nil, err
		}
		m.SentAt = sentAt
		out = append(out, m)
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, rows.Err()
}
