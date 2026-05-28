package employees

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/predicta/predicta/internal/domain/entity"
)

// FileRepository — roster команды из JSON (Jira external_id + Telegram nick).
type FileRepository struct {
	teamID     string
	byID       map[string]entity.Employee
	byNick     map[string]entity.Employee
	byExternal map[string]entity.Employee
}

type teamConfig struct {
	TeamID  string         `json:"team_id"`
	Members []memberConfig `json:"members"`
}

type memberConfig struct {
	ID           string `json:"id"`
	ExternalID   string `json:"external_id"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	TelegramNick string `json:"telegram_nick"`
}

func LoadFile(path string) (*FileRepository, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read employees config: %w", err)
	}

	var cfg teamConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse employees config: %w", err)
	}
	if cfg.TeamID == "" {
		cfg.TeamID = "backend"
	}
	if len(cfg.Members) == 0 {
		return nil, fmt.Errorf("employees config: no members")
	}

	repo := &FileRepository{
		teamID:     cfg.TeamID,
		byID:       make(map[string]entity.Employee),
		byNick:     make(map[string]entity.Employee),
		byExternal: make(map[string]entity.Employee),
	}

	for _, m := range cfg.Members {
		if m.ID == "" {
			return nil, fmt.Errorf("employee %q: id is required", m.Name)
		}
		ext := m.ExternalID
		if ext == "" {
			ext = m.ID
		}
		e := entity.Employee{
			ID:           m.ID,
			ExternalID:   ext,
			Name:         m.Name,
			Role:         m.Role,
			TelegramNick: strings.TrimPrefix(strings.TrimSpace(m.TelegramNick), "@"),
		}
		repo.byID[e.ID] = e
		repo.byExternal[e.ExternalID] = e
		if e.TelegramNick != "" {
			repo.byNick[normalizeNick(e.TelegramNick)] = e
		}
	}

	return repo, nil
}

func (r *FileRepository) TeamID() string {
	return r.teamID
}

func (r *FileRepository) ListByTeam(_ context.Context, teamID string) ([]entity.Employee, error) {
	if teamID != r.teamID {
		return nil, fmt.Errorf("team not found: %s", teamID)
	}
	out := make([]entity.Employee, 0, len(r.byID))
	for _, e := range r.byID {
		out = append(out, e)
	}
	return out, nil
}

func (r *FileRepository) GetByID(_ context.Context, id string) (*entity.Employee, error) {
	if e, ok := r.byID[id]; ok {
		copy := e
		return &copy, nil
	}
	if e, ok := r.byExternal[id]; ok {
		copy := e
		return &copy, nil
	}
	return nil, fmt.Errorf("employee not found: %s", id)
}

func (r *FileRepository) GetByTelegramNick(_ context.Context, nick string) (*entity.Employee, error) {
	if e, ok := r.byNick[normalizeNick(nick)]; ok {
		copy := e
		return &copy, nil
	}
	return nil, fmt.Errorf("employee not found for nick: %s", nick)
}

func normalizeNick(nick string) string {
	nick = strings.TrimSpace(nick)
	nick = strings.TrimPrefix(nick, "@")
	return strings.ToLower(nick)
}
