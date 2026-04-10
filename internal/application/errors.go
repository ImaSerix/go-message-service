package application

import (
	"errors"
	"fmt"

	"github.com/ImaSerix/go-message-service/internal/domain/user"
)

var (
	ErrAlreadyExists = errors.New("already exists")
)

type ErrUsersNotExists struct {
	Msg   string
	Users []user.UserID
}

func (err *ErrUsersNotExists) Error() string {
	return fmt.Sprintf("%s %v", err.Msg, err.Users)
}
