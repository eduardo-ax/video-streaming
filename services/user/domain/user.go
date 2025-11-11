package domain

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	AccessToken string      `json:"access_token"`
	User        UserPayload `json:"user"`
}

type UserInterface interface {
	CreateUser(ctx context.Context, name string, email string, plan int8, pass string) error
	UserLogin(ctx context.Context, email string, password string) (*LoginUserRes, error)
}

type UserManager struct {
	db    Storage
	token TokenInterface
}

func NewUserManager(db Storage, token TokenInterface) *UserManager {
	return &UserManager{
		db:    db,
		token: token,
	}
}

func (u *UserManager) CreateUser(ctx context.Context, name string, email string, plan int8, password string) error {

	cryptPassword, err := HashPassword(password)
	if err != nil {
		return err
	}
	_, err = u.db.Persist(ctx, name, email, cryptPassword, plan)

	if err != nil {
		return err
	}
	return nil

}

func (u *UserManager) UserLogin(ctx context.Context, email string, password string) (*LoginUserRes, error) {
	user, err := u.db.GetUser(ctx, email)

	if err != nil {
		return nil, err
	}
	if !CheckPassword(password, user.Password) {
		return nil, fmt.Errorf("password incorrect")
	}

	acessToken, err := u.token.CreateToken(user.ID, email, user.Plan, 15*time.Minute)

	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}
	return &LoginUserRes{
		AccessToken: acessToken,
		User: UserPayload{
			ID:    user.ID,
			Email: email,
			Plan:  user.Plan,
		},
	}, nil

}

func HashPassword(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(pass, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}

type Storage interface {
	Persist(ctx context.Context, name string, email string, password string, plan int8) (string, error)
	GetUser(ctx context.Context, email string) (*UserAuthData, error)
}

type TokenInterface interface {
	CreateToken(id string, email string, plan int8, duration time.Duration) (string, error)
	VerifyToken(tokenStr string) (*UserPayload, error)
}
