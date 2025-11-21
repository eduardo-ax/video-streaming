package api

import "time"

type UserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
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

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type LoginSuccessResponse struct {
	SessionID          string      `json:"session_id"`
	AccessToken        string      `json:"access_token"`
	AccessTokenExpires time.Time   `json:"access_token_expires_at"`
	User               interface{} `json:"user"`
	Message            string      `json:"message"`
}

type CreateUserResponse struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Plan    int32  `json:"plan"`
	Message string `json:"message"`
}
