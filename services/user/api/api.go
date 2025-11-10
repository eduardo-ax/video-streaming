package api

import (
	"fmt"
	"net/http"

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

type UserHandler struct {
	user domain.UserInterface
}

func NewUserHander(user domain.UserInterface) *UserHandler {
	return &UserHandler{
		user: user,
	}
}

func (u *UserHandler) Register(e *echo.Group) {
	e.POST("/user", u.CreateUserHandler)
	e.POST("/login", u.LoginHandler)
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

	userID, err := u.user.UserLogin(ctx, userLogin.Email, userLogin.Password)

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "incorrect credentials")
	}

	fmt.Println(userID)
	return echo.NewHTTPError(http.StatusOK, "login successfully")
}
