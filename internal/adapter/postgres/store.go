package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/predicta/predicta/internal/domain/entity"
	_ "github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) ListByTeam(ctx context.Context, teamID string) ([]entity.Employee, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, external_id, name, role, telegram_nick
		FROM employees
		WHERE team_id = $1
		ORDER BY name`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.Employee, 0)
	for rows.Next() {
		var e entity.Employee
		if err := rows.Scan(&e.ID, &e.ExternalID, &e.Name, &e.Role, &e.TelegramNick); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) GetByID(ctx context.Context, id string) (*entity.Employee, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, external_id, name, role, telegram_nick
		FROM employees WHERE id = $1`, id)

	var e entity.Employee
	if err := row.Scan(&e.ID, &e.ExternalID, &e.Name, &e.Role, &e.TelegramNick); err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Store) GetByTelegramNick(ctx context.Context, nick string) (*entity.Employee, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, external_id, name, role, telegram_nick
		FROM employees WHERE telegram_nick = $1`, nick)

	var e entity.Employee
	if err := row.Scan(&e.ID, &e.ExternalID, &e.Name, &e.Role, &e.TelegramNick); err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Store) Save(ctx context.Context, msg entity.ChatMessage) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO chat_messages (employee_id, message_text, sent_at)
		VALUES ($1, $2, $3)`,
		msg.EmployeeID, msg.Text, msg.SentAt,
	)
	return err
}

func (s *Store) ListRecentByEmployee(ctx context.Context, employeeID string, limit int) ([]entity.ChatMessage, error) {
	rows, err := s.db.QueryContext(ctx, `
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
