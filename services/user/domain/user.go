package domain

type User struct {
	ID             string
	Name           string
	Email          string
	Password       string
	Plan           string
	VideosQuantity int8
	UserVideosID   []int
}

type UserInterface interface {
	Created(name string, email string, plan string)
}

type UserManager struct {
}

func NewUserManager() *UserManager {
	return &UserManager{}
}

func (*UserManager) Created(name string, email string, plan string) {

}
