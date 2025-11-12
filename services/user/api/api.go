package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eduardo-ax/video-streaming/services/user/domain"
	"github.com/eduardo-ax/video-streaming/services/user/token"
	"github.com/labstack/echo/v4"
)

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

type RenewAccessTokenReq struct {
	RefreshToken string `json:"refresh_token"`
}

type UserHandler struct {
	user domain.UserInterface
}

type RenewAccessTokenRes struct {
	AccessToken         string    `json:"access_token"`
	AcessTokenExpiresAt time.Time `json:"acess_token_expires_at"`
}

func NewUserHander(user domain.UserInterface) *UserHandler {
	return &UserHandler{
		user: user,
	}
}

const ContextUserID = "userID"

func (u *UserHandler) AuthMiddleware(tokenMaker *token.JWTMaker) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing authorization header")
			}

			fields := strings.Fields(authHeader)
			if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization format")
			}

			accessToken := fields[1]
			claims, err := tokenMaker.VerifyToken(accessToken)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired access token")
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

	protected.POST("/logout/:id", u.LogoutHandler)
	protected.POST("/revoke/:id", u.RevokeTokenHandler)
}

func (u *UserHandler) CreateUserHandler(c echo.Context) error {
	ctx := c.Request().Context()
	user := &UserRequest{}

	if err := c.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	err := u.user.CreateUser(ctx, user.Name, user.Email, user.Plan, user.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Create User error: %s", err))
	}

	return echo.NewHTTPError(http.StatusCreated, "user created successfully")
}

func (u *UserHandler) DeleteUserHandler(c echo.Context) error {
	ctx := c.Request().Context()

	loggedInUserID, ok := c.Get(ContextUserID).(string)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "User ID not available in context")

	}

	err := u.user.DeleteUser(ctx, loggedInUserID)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error deleting user: %s", err))
	}
	return echo.NewHTTPError(http.StatusOK, "User deleted successfully")
}

func (u *UserHandler) UpdateUserHandler(c echo.Context) error {
	ctx := c.Request().Context()
	user := &UserRequest{}

	if err := c.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	err := u.user.CreateUser(ctx, user.Name, user.Email, user.Plan, user.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Create User error: %s", err))
	}

	return echo.NewHTTPError(http.StatusCreated, "user created successfully")

}

func (u *UserHandler) LoginHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userLogin := &LoginUserRequest{}

	if err := c.Bind(userLogin); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	userClaims, err := u.user.UserLogin(ctx, userLogin.Email, userLogin.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "incorrect credentials")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "login successfully",
		"user":    userClaims,
	})
}

func (u *UserHandler) LogoutHandler(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	err := u.user.UserLogout(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error logout user")
	}
	return echo.NewHTTPError(http.StatusOK, "logout successfully")
}

func (u *UserHandler) RenewTokenHandler(c echo.Context) error {
	ctx := c.Request().Context()
	refreshToken := &RenewAccessTokenReq{}
	if err := c.Bind(refreshToken); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	req, err := u.user.RenewAccessToken(ctx, refreshToken.RefreshToken)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error renew  token")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "renew successfully",
		"user":    req,
	})

}

func (u *UserHandler) RevokeTokenHandler(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	err := u.user.RevokeSession(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error logout user")
	}
	return echo.NewHTTPError(http.StatusOK, "session revoke successfuly")
}
