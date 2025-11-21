package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eduardo-ax/video-streaming/services/user/domain"
	"github.com/eduardo-ax/video-streaming/services/user/token"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	user    domain.UserInterface
	metrics Metrics
}

func NewUserHander(user domain.UserInterface, metrics Metrics) *UserHandler {
	return &UserHandler{
		user:    user,
		metrics: metrics,
	}
}

func JSONError(c echo.Context, status int, message string) error {
	return c.JSON(status, map[string]string{"error": message})
}

func JSONSucess(c echo.Context, status int, message string) error {
	return c.JSON(status, map[string]string{"message": message})
}

func SetRefreshTokenCookie(c echo.Context, refreshToken string, expiresAt time.Time) {
	cookie := new(http.Cookie)
	cookie.Name = "refresh_token"
	cookie.Value = refreshToken
	cookie.Expires = expiresAt
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.Path = "/"
	cookie.SameSite = http.SameSiteLaxMode

	c.SetCookie(cookie)
}

func ClearRefreshTokenCookie(c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = "refresh_token"
	cookie.Value = ""
	cookie.MaxAge = -1
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode
	c.SetCookie(cookie)
}

const ContextUserID = "userID"
const ContextSessionID = "sessionID"

func (u *UserHandler) AuthMiddleware(tokenMaker *token.JWTMaker) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return JSONError(c, http.StatusUnauthorized, "missing authorization header")
			}

			fields := strings.Fields(authHeader)
			if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" {
				return JSONError(c, http.StatusUnauthorized, "invalid authorization format")
			}

			accessToken := fields[1]
			claims, err := tokenMaker.VerifyToken(accessToken)
			if err != nil {
				return JSONError(c, http.StatusUnauthorized, "invalid or expired access token")
			}
			c.Set(ContextUserID, claims.ID)
			c.Set(ContextSessionID, claims.RegisteredClaims.ID)
			return next(c)
		}
	}
}

func (u *UserHandler) Register(e *echo.Group, tokenMaker *token.JWTMaker) {
	e.POST("/user", u.CreateUserHandler)
	e.POST("/login", u.LoginHandler)
	e.POST("/renew", u.RenewTokenHandler)

	protected := e.Group("")
	protected.Use(u.AuthMiddleware(tokenMaker))

	protected.PUT("/user", u.UpdateUserHandler)
	protected.DELETE("/user", u.DeleteUserHandler)

	protected.POST("/logout/", u.LogoutHandler)
	protected.POST("/revoke/:id", u.RevokeTokenHandler)
}

// CreateUserHandler godoc
// @Summary Create user
// @Description Create user
// @Tags User
// @Accept json
// @Produce json
// @Param login body UserRequest true "Create user credentials"
// @Success 200 {object} MessageResponse
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /v1/user [post]
func (u *UserHandler) CreateUserHandler(c echo.Context) error {
	ctx := c.Request().Context()
	user := &UserRequest{}

	if err := c.Bind(user); err != nil {
		return JSONError(c, http.StatusBadRequest, "invalid request body")
	}

	err := u.user.CreateUser(ctx, user.Name, user.Email, user.Password)
	if err != nil {
		return JSONError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create user: %s", err))
	}
	u.metrics.IncUsersCreated()

	return JSONSucess(c, http.StatusCreated, "user created successfully")
}

// DeleteUserHandler godoc
// @Summary Delete user
// @Description Delete user
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /v1/user [delete]
func (u *UserHandler) DeleteUserHandler(c echo.Context) error {
	ctx := c.Request().Context()

	loggedInUserID, ok := c.Get(ContextUserID).(string)
	if !ok || loggedInUserID == "" {
		return JSONError(c, http.StatusUnauthorized, "user ID not available in context")
	}

	err := u.user.DeleteUser(ctx, loggedInUserID)
	if err != nil {
		return JSONError(c, http.StatusInternalServerError, fmt.Sprintf("failed to delete user: %s", err))
	}
	return JSONSucess(c, http.StatusNoContent, "user deleted successfully")
}

// UpdateUserHandler godoc
// @Summary Update user
// @Description Update user credentials
// @Tags User
// @Accept json
// @Produce json
// @Param login body UpdateUserRequest true "Update user credentials"
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /v1/user [put]
func (u *UserHandler) UpdateUserHandler(c echo.Context) error {
	ctx := c.Request().Context()

	loggedInUserID, ok := c.Get(ContextUserID).(string)
	if !ok || loggedInUserID == "" {
		return JSONError(c, http.StatusUnauthorized, "user ID not available in context")
	}

	req := &UpdateUserRequest{}
	if err := c.Bind(&req); err != nil {
		return JSONError(c, http.StatusBadRequest, "invalid request body format.")
	}

	if req.Name == "" && req.Email == nil && req.Password == nil {
		return JSONError(c, http.StatusBadRequest, "no fields provided")
	}

	err := u.user.UpdateUser(ctx, loggedInUserID, req.Name, req.Email, req.Password)
	if err != nil {
		return JSONError(c, http.StatusInternalServerError, fmt.Sprintf("failed to update user: %s", err))
	}
	return JSONSucess(c, http.StatusOK, "user updated successfully")
}

// LoginHandler godoc
// @Summary User login
// @Description Authenticate user and return tokens
// @Tags User
// @Accept json
// @Produce json
// @Param login body LoginUserRequest true "User login credentials"
// @Success 200 {object} LoginSuccessResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Invalid email or password"
// @Router /v1/login [post]
func (u *UserHandler) LoginHandler(c echo.Context) error {
	ctx := c.Request().Context()
	start := time.Now()
	userLogin := &LoginUserRequest{}

	if err := c.Bind(userLogin); err != nil {
		return JSONError(c, http.StatusBadRequest, "invalid request body format")
	}

	userClaims, err := u.user.UserLogin(ctx, userLogin.Email, userLogin.Password)
	if err != nil {
		return JSONError(c, http.StatusUnauthorized, "incorrect credentials")
	}

	SetRefreshTokenCookie(c, userClaims.RefreshToken, userClaims.RefreshTokenExpiresAt)

	u.metrics.IncLoginSuccess()
	u.metrics.ObserveLoginDuration(time.Since(start).Seconds())
	resp := LoginSuccessResponse{
		SessionID:          userClaims.SessionID,
		AccessToken:        userClaims.AccessToken,
		AccessTokenExpires: userClaims.AccessTokenExpiresAt,
		User:               userClaims.User,
		Message:            "login successfully",
	}
	return c.JSON(http.StatusOK, resp)
}

// LogoutHandler godoc
// @Summary Logout user
// @Description Deletes user session and clears refresh token cookie
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /v1/logout/ [post]
func (u *UserHandler) LogoutHandler(c echo.Context) error {
	ctx := c.Request().Context()

	sessionID, ok := c.Get(ContextSessionID).(string)
	if !ok || sessionID == "" {
		return JSONError(c, http.StatusUnauthorized, "invalid authorization format")
	}

	err := u.user.UserLogout(ctx, sessionID)
	if err != nil {
		return JSONError(c, http.StatusInternalServerError, fmt.Sprintf("failed to logout user: %s", err))
	}

	ClearRefreshTokenCookie(c)

	return JSONSucess(c, http.StatusOK, "logout successfully")
}

// RenewTokenHandler godoc
// @Summary Renew access token
// @Description Renew access token usinng refresh token
// @Tags Token
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /v1/renew [post]
func (u *UserHandler) RenewTokenHandler(c echo.Context) error {
	ctx := c.Request().Context()
	start := time.Now()

	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return JSONError(c, http.StatusUnauthorized, "refresh token required")
		}
		return JSONError(c, http.StatusUnauthorized, "invalid request")
	}

	refreshTokenValue := cookie.Value
	renewResponse, err := u.user.RenewAccessToken(ctx, refreshTokenValue)
	if err != nil {
		return JSONError(c, http.StatusInternalServerError, "failed to renew token")
	}

	u.metrics.IncSessionRenew()
	u.metrics.ObserveSessionRenewDuration(time.Since(start).Seconds())
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":      "renew successfully",
		"access_token": renewResponse.AccessToken,
	})
}

// RevokeTokenHandler godoc
// @Summary Revoke refresh token
// @Description Revoke refresh token using session id
// @Tags Token
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /v1/revoke/:id [post]
func (u *UserHandler) RevokeTokenHandler(c echo.Context) error {
	ctx := c.Request().Context()

	SessionID, ok := c.Get(ContextSessionID).(string)
	if !ok || SessionID == "" {
		return JSONError(c, http.StatusUnauthorized, "user ID not available in context")
	}

	err := u.user.RevokeSession(ctx, SessionID)
	if err != nil {
		return JSONError(c, http.StatusInternalServerError, fmt.Sprintf("failed to revoke session %s", err))
	}
	return JSONSucess(c, http.StatusOK, "session revoked successfully")
}

type Metrics interface {
	IncUsersCreated()
	IncLoginSuccess()
	IncLoginError()
	ObserveLoginDuration(seconds float64)
	IncSessionRenew()
	ObserveSessionRenewDuration(seconds float64)
}
