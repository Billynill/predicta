package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
	"golang.org/x/crypto/bcrypt"
)

type AuthRegisterer interface {
	Register(ctx context.Context, cmd RegisterManagerCommand) error
}

type AuthLogin interface {
	Login(ctx context.Context, email, password string) (string, error)
}

type AuthProfileGetter interface {
	GetProfile(ctx context.Context, managerID string) (entity.ManagerProfile, error)
}

type RegisterManagerCommand struct {
	FirstName    string
	LastName     string
	Email        string
	Password     string
	TelegramNick string
	Phone        string
	AvatarURL    string
}

type AuthService struct {
	managers  port.ManagerRepository
	tokens    port.TokenIssuer
	employees port.EmployeeRepository
	jiraSetup port.JiraSetup
	teamID    string
}

func NewAuthService(
	managers port.ManagerRepository,
	tokens port.TokenIssuer,
	employees port.EmployeeRepository,
	jiraSetup port.JiraSetup,
	teamID string,
) *AuthService {
	return &AuthService{
		managers:  managers,
		tokens:    tokens,
		employees: employees,
		jiraSetup: jiraSetup,
		teamID:    resolveTeamID(teamID),
	}
}

func (s *AuthService) Register(ctx context.Context, cmd RegisterManagerCommand) error {
	if err := validateRegister(cmd); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	manager := entity.Manager{
		ID:           uuid.NewString(),
		FirstName:    strings.TrimSpace(cmd.FirstName),
		LastName:     strings.TrimSpace(cmd.LastName),
		Email:        strings.ToLower(strings.TrimSpace(cmd.Email)),
		PasswordHash: string(hash),
		TelegramNick: strings.TrimPrefix(strings.TrimSpace(cmd.TelegramNick), "@"),
		Phone:        strings.TrimSpace(cmd.Phone),
		AvatarURL:    strings.TrimSpace(cmd.AvatarURL),
		CreatedAt:    time.Now().UTC(),
	}

	return s.managers.Create(ctx, manager)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	manager, err := s.managers.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if errors.Is(err, port.ErrManagerNotFound) {
			return "", port.ErrInvalidCredentials
		}
		return "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(manager.PasswordHash), []byte(password)) != nil {
		return "", port.ErrInvalidCredentials
	}
	return s.tokens.Issue(manager.ID)
}

func (s *AuthService) GetProfile(ctx context.Context, managerID string) (entity.ManagerProfile, error) {
	manager, err := s.managers.GetByID(ctx, managerID)
	if err != nil {
		return entity.ManagerProfile{}, err
	}

	subordinates, err := s.employees.ListByTeam(ctx, s.teamID)
	if err != nil {
		return entity.ManagerProfile{}, err
	}

	profile := entity.ManagerProfile{
		FirstName:         manager.FirstName,
		LastName:          manager.LastName,
		Email:             manager.Email,
		TelegramNick:      manager.TelegramNick,
		Phone:             manager.Phone,
		SubordinatesCount: len(subordinates),
		AvatarURL:         manager.AvatarURL,
	}

	if profile.AvatarURL == "" && s.jiraSetup != nil {
		jiraUser, err := s.jiraSetup.WhoAmI(ctx)
		if err == nil && jiraUser != nil {
			profile.AvatarURL = jiraUser.AvatarURL
			profile.JiraDisplayName = jiraUser.DisplayName
			profile.JiraEmail = jiraUser.Email
		}
	} else if s.jiraSetup != nil {
		jiraUser, err := s.jiraSetup.WhoAmI(ctx)
		if err == nil && jiraUser != nil {
			profile.JiraDisplayName = jiraUser.DisplayName
			profile.JiraEmail = jiraUser.Email
		}
	}

	return profile, nil
}

func validateRegister(cmd RegisterManagerCommand) error {
	if strings.TrimSpace(cmd.FirstName) == "" {
		return fmt.Errorf("first_name is required")
	}
	if strings.TrimSpace(cmd.LastName) == "" {
		return fmt.Errorf("last_name is required")
	}
	if strings.TrimSpace(cmd.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if len(cmd.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	if strings.TrimSpace(cmd.TelegramNick) == "" {
		return fmt.Errorf("telegram_nick is required")
	}
	if strings.TrimSpace(cmd.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	return nil
}
