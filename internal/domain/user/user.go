package user

type User struct {
	id       UserID
	username Username
}

func NewUser(id UserID, username Username) *User {
	return &User{
		id:       id,
		username: username,
	}
}
