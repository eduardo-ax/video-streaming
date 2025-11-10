package domain

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string
	Name     string
	Email    string
	Password string
	Plan     string
}

type UserAuthData struct {
	ID       string
	Password string
}

type UserInterface interface {
	CreateUser(ctx context.Context, name string, email string, plan int8, pass string) error
	UserLogin(ctx context.Context, email string, password string) (string, error)
}

type UserManager struct {
	db Storage
}

func NewUserManager(db Storage) *UserManager {
	return &UserManager{
		db: db,
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

func (u *UserManager) UserLogin(ctx context.Context, email string, password string) (string, error) {
	user, err := u.db.GetUser(ctx, email)

	if err != nil {
		return "", err
	}
	if CheckPassword(password, user.Password) {
		return user.ID, nil
	} else {
		return "", fmt.Errorf("password incorrect")
	}
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
