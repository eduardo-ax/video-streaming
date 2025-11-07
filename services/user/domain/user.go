package domain

import (
	"context"
	"fmt"
)

type User struct {
	ID       string
	Name     string
	Email    string
	Password string
	Plan     string
}

type UserInterface interface {
	Created(ctx context.Context, name string, email string, plan int8, pass string) error
}

type UserManager struct {
	db Storage
}

func NewUserManager(db Storage) *UserManager {
	return &UserManager{
		db: db,
	}
}

func (u *UserManager) Created(ctx context.Context, name string, email string, plan int8, password string) error {

	fmt.Println(name)
	fmt.Println(email)
	fmt.Println(plan)
	fmt.Println(password)
	_, err := u.db.Persist(ctx, name, email, password, plan)

	if err != nil {
		return err
	}
	return nil
}

type Storage interface {
	Persist(ctx context.Context, name string, email string, password string, plan int8) (int, error)
}
