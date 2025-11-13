package api

import "time"

type UserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Plan     int8   `json:"plan"`
	Password string `json:"password"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Name     string  `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
}

type RenewAccessTokenRes struct {
	AccessToken         string    `json:"access_token"`
	AcessTokenExpiresAt time.Time `json:"acess_token_expires_at"`
}
