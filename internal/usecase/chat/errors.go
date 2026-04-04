package chat

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrInvalidName   = errors.New("invalid name")
	ErrUserNotExists = errors.New("user not exists")
	ErrAlreadyExists = errors.New("already exists")
)

type ErrUsersNotExists struct {
	msg   string
	users []uuid.UUID
}

func (err *ErrUsersNotExists) Error() string {
	return fmt.Sprintf("%s %v", err.msg, err.users)
}
