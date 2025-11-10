package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserClaims struct {
	ID    string `json:"id`
	Email string `json:"email"`
	Plan  int8   `json:"plan"`
	jwt.RegisteredClaims
}

func NewUserClaims(id string, email string, plan int8, duration time.Duration) (*UserClaims, error) {
	tokenID, err := uuid.NewRandom()

	if err != nil {
		return nil, fmt.Errorf("error generating token ID: %w", err)
	}

	return &UserClaims{
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
