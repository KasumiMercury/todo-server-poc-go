package user

type User struct {
	id       string
	username string
}

func NewUser(id, username string) *User {
	return &User{
		id:       id,
		username: username,
	}
}

func (u *User) ID() string {
	return u.id
}

func (u *User) Username() string {
	return u.username
}
