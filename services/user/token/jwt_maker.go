package token

import (
	"fmt"
	"time"

	"github.com/eduardo-ax/video-streaming/services/user/domain"
	"github.com/golang-jwt/jwt/v5"
)

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) *JWTMaker {
	return &JWTMaker{secretKey: secretKey}
}

func (m *JWTMaker) CreateToken(id string, email string, plan int8, sessionID string, duration time.Duration) (string, *domain.UserClaims, error) {
	claims, err := NewUserClaims(id, email, plan, sessionID, duration)

	if err != nil {
		return "", nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", nil, fmt.Errorf("error signing token: %w", err)
	}
	return tokenStr, claims, nil
}

func (m JWTMaker) VerifyToken(tokenStr string) (*domain.UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &domain.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("invalid token signing method")
		}

		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsin token %w", err)
	}
	claims, ok := token.Claims.(*domain.UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
