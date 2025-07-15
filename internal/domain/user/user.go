package user

type User struct {
	id       string
	username string
}

func NewUser() *User {
	return &User{}
}

func (u *User) ID() string {
	return u.id
}

func (u *User) Username() string {
	return u.username
}
