package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/predicta/predicta/internal/domain/port"
)

const tokenTTL = 7 * 24 * time.Hour

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

type claims struct {
	ManagerID string `json:"mid"`
	jwt.RegisteredClaims
}

func (s *JWTService) Issue(managerID string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		ManagerID: managerID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   managerID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	})
	return token.SignedString(s.secret)
}

func (s *JWTService) Parse(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return "", err
	}

	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid || c.ManagerID == "" {
		return "", errors.New("invalid token")
	}
	return c.ManagerID, nil
}

var _ port.TokenIssuer = (*JWTService)(nil)
