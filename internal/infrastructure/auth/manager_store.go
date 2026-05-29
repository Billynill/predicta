package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
)

var _ port.ManagerRepository = (*FileManagerStore)(nil)

// FileManagerStore — один начальник в JSON-файле.
type FileManagerStore struct {
	path string
	mu   sync.RWMutex
}

type managerRecord struct {
	ID           string `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	TelegramNick string `json:"telegram_nick"`
	Phone        string `json:"phone"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	CreatedAt    string `json:"created_at"`
}

func NewFileManagerStore(path string) *FileManagerStore {
	return &FileManagerStore{path: path}
}

func (s *FileManagerStore) Exists(_ context.Context) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, err := s.readLocked()
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *FileManagerStore) Create(_ context.Context, m entity.Manager) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.readLocked(); err == nil {
		return port.ErrManagerExists
	} else if !os.IsNotExist(err) {
		return err
	}

	rec := managerRecord{
		ID:           m.ID,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		Email:        strings.ToLower(strings.TrimSpace(m.Email)),
		PasswordHash: m.PasswordHash,
		TelegramNick: normalizeNick(m.TelegramNick),
		Phone:        strings.TrimSpace(m.Phone),
		AvatarURL:    strings.TrimSpace(m.AvatarURL),
		CreatedAt:    m.CreatedAt.UTC().Format(timeRFC3339),
	}
	return s.writeLocked(rec)
}

func (s *FileManagerStore) GetByEmail(_ context.Context, email string) (*entity.Manager, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rec, err := s.readLocked()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, port.ErrManagerNotFound
		}
		return nil, err
	}
	if !strings.EqualFold(rec.Email, email) {
		return nil, port.ErrManagerNotFound
	}
	return recordToEntity(rec)
}

func (s *FileManagerStore) GetByID(_ context.Context, id string) (*entity.Manager, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rec, err := s.readLocked()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, port.ErrManagerNotFound
		}
		return nil, err
	}
	if rec.ID != id {
		return nil, port.ErrManagerNotFound
	}
	return recordToEntity(rec)
}

func (s *FileManagerStore) readLocked() (managerRecord, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return managerRecord{}, err
	}
	var rec managerRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return managerRecord{}, fmt.Errorf("parse manager store: %w", err)
	}
	return rec, nil
}

func (s *FileManagerStore) writeLocked(rec managerRecord) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

func recordToEntity(rec managerRecord) (*entity.Manager, error) {
	created, _ := timeParse(rec.CreatedAt)
	return &entity.Manager{
		ID:           rec.ID,
		FirstName:    rec.FirstName,
		LastName:     rec.LastName,
		Email:        rec.Email,
		PasswordHash: rec.PasswordHash,
		TelegramNick: rec.TelegramNick,
		Phone:        rec.Phone,
		AvatarURL:    rec.AvatarURL,
		CreatedAt:    created,
	}, nil
}

func normalizeNick(nick string) string {
	return strings.TrimPrefix(strings.TrimSpace(nick), "@")
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"

func timeParse(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(timeRFC3339, s)
}
