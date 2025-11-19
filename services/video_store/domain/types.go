package domain

import "github.com/golang-jwt/jwt/v5"

type UserClaims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Plan  int8   `json:"plan"`
	jwt.RegisteredClaims
}
