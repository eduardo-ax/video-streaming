package token

import (
	"time"

	"github.com/eduardo-ax/video-streaming/services/user/domain"
	"github.com/golang-jwt/jwt/v5"
)

func NewUserClaims(id string, email string, plan int8, sessionID string, duration time.Duration) (*domain.UserClaims, error) {

	return &domain.UserClaims{
		Email: email,
		ID:    id,
		Plan:  plan,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionID,
			Subject:   email,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}, nil
}
