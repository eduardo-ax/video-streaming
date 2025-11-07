package api

import (
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
}

func (u *UserHandler) CreateUserHandler(c echo.Context) error {
	ctx := c.Request().Context()
	user := &UserRequest{}

	if err := c.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	err := u.user.Created(ctx, user.Name, user.Email, user.Plan, user.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	return echo.NewHTTPError(http.StatusCreated, "user created successfully")

}
