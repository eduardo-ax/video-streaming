package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eduardo-ax/video-streaming/services/user/domain"
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

func (u *UserHandler) Register(e *echo.Group) {
	e.POST("/user", u.CreateUserHandler)
	e.POST("/login", u.LoginHandler)
	e.POST("/logout/:id", u.LogoutHandler)
	e.POST("/renew", u.RenewTokenHandler)
	e.POST("/revoke/:id", u.RevokeTokenHandler)
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
