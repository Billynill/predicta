package port

import (
	"context"
	"errors"

	"github.com/predicta/predicta/internal/domain/entity"
)

var (
	ErrManagerExists       = errors.New("manager already registered")
	ErrManagerNotFound     = errors.New("manager not found")
	ErrInvalidCredentials  = errors.New("invalid email or password")
)

type ManagerRepository interface {
	Exists(ctx context.Context) (bool, error)
	Create(ctx context.Context, manager entity.Manager) error
	GetByEmail(ctx context.Context, email string) (*entity.Manager, error)
	GetByID(ctx context.Context, id string) (*entity.Manager, error)
}

type TokenIssuer interface {
	Issue(managerID string) (string, error)
	Parse(token string) (managerID string, err error)
}
