package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID       string
	Name     string
	Email    string
	Password string
	Plan     string
}

type UserPayload struct {
	ID    string
	Email string
	Plan  int8
}

type UserAuthData struct {
	ID       string
	Password string
	Plan     int8
}

type LoginUserRes struct {
	SessionID             string      `json:"session_id"`
	AccessToken           string      `json:"access_token"`
	RefreshToken          string      `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time   `json:"acess_token_expires_at"`
	RefreshTokenExpiresAt time.Time   `json:"refresh_token_expires_at"`
	User                  UserPayload `json:"user"`
}

type Session struct {
	ID           string
	Email        string
	RefreshToken string
	IsRevoked    bool
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

type RenewAcessTokenReq struct {
	RefreshToken string `json:"refresh_token"`
}

type RenewAccessTokenRes struct {
	AccessToken         string    `json:"access_token"`
	AcessTokenExpiresAt time.Time `json:"acess_token_expires_at"`
}

type UserClaims struct {
	ID    string `json:"id`
	Email string `json:"email"`
	Plan  int8   `json:"plan"`
	jwt.RegisteredClaims
}

type UserManager struct {
	db    Storage
	token TokenInterface
}
