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
	user domain.UserInterface
}

func NewUserHander(user domain.UserInterface) *UserHandler {
	return &UserHandler{
		user: user,
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

const ContextUserID = "userID"

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
			return next(c)
		}
	}
}

func (u *UserHandler) Register(g *echo.Group, tokenMaker *token.JWTMaker) {
	g.POST("/user", u.CreateUserHandler)
	g.POST("/login", u.LoginHandler)
	g.POST("/renew", u.RenewTokenHandler)

	protected := g.Group("")
	protected.Use(u.AuthMiddleware(tokenMaker))

	protected.PUT("/user", u.UpdateUserHandler)
	protected.DELETE("/user", u.DeleteUserHandler)

	protected.POST("/logout/", u.LogoutHandler)
	protected.POST("/revoke/:id", u.RevokeTokenHandler)
}

func (u *UserHandler) CreateUserHandler(c echo.Context) error {
	ctx := c.Request().Context()
	user := &UserRequest{}

	if err := c.Bind(user); err != nil {
		return JSONError(c, http.StatusBadRequest, "invalid request body")
	}

	err := u.user.CreateUser(ctx, user.Name, user.Email, user.Plan, user.Password)
	if err != nil {
		return JSONError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create user: %s", err))
	}
	return JSONSucess(c, http.StatusCreated, "user created successfully")
}

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

func (u *UserHandler) LoginHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userLogin := &LoginUserRequest{}

	if err := c.Bind(userLogin); err != nil {
		return JSONError(c, http.StatusBadRequest, "invalid request body format")
	}

	userClaims, err := u.user.UserLogin(ctx, userLogin.Email, userLogin.Password)
	if err != nil {
		return JSONError(c, http.StatusUnauthorized, "incorrect credentials")
	}

	refreshToken := userClaims.RefreshToken
	expiresAtStr := userClaims.RefreshTokenExpiresAt

	SetRefreshTokenCookie(c, refreshToken, expiresAtStr)

	responseBody := map[string]interface{}{
		"session_id":             userClaims.SessionID,
		"access_token":           userClaims.AccessToken,
		"acess_token_expires_at": userClaims.AccessTokenExpiresAt,
		"user":                   userClaims.User,
		"message":                "login successfully",
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "login successfully",
		"user":    responseBody,
	})
}

func (u *UserHandler) LogoutHandler(c echo.Context) error {
	ctx := c.Request().Context()

	loggedInUserID, ok := c.Get(ContextUserID).(string)
	if !ok || loggedInUserID == "" {
		return JSONError(c, http.StatusUnauthorized, "user ID not available in context")
	}

	err := u.user.UserLogout(ctx, ContextUserID)
	if err != nil {
		return JSONError(c, http.StatusInternalServerError, fmt.Sprintf("failed to logout user %s", err))
	}
	return JSONSucess(c, http.StatusOK, "logout successfully")
}

func (u *UserHandler) RenewTokenHandler(c echo.Context) error {
	ctx := c.Request().Context()

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
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":      "renew successfully",
		"access_token": renewResponse.AccessToken,
	})
}

func (u *UserHandler) RevokeTokenHandler(c echo.Context) error {
	ctx := c.Request().Context()

	loggedInUserID, ok := c.Get(ContextUserID).(string)
	if !ok || loggedInUserID == "" {
		return JSONError(c, http.StatusUnauthorized, "user ID not available in context")
	}

	id := c.Param("id")
	if id == "" {
		return JSONError(c, http.StatusBadRequest, "invalid request body")
	}

	err := u.user.RevokeSession(ctx, id)
	if err != nil {
		return JSONError(c, http.StatusInternalServerError, fmt.Sprintf("failed to revoke session %s", err))
	}
	return JSONSucess(c, http.StatusOK, "session revoked successfully")
}
