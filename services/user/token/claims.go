package token

import (
	"fmt"
	"time"

	"github.com/eduardo-ax/video-streaming/services/user/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func NewUserClaims(id string, email string, plan int8, duration time.Duration) (*domain.UserClaims, error) {
	tokenID, err := uuid.NewRandom()

	if err != nil {
		return nil, fmt.Errorf("error generating token ID: %w", err)
	}

	return &domain.UserClaims{
		Email: email,
		ID:    id,
		Plan:  plan,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			Subject:   email,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}, nil
}
