package chat

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrInvalidName            = errors.New("invalid name")
	ErrAlreadyExists          = errors.New("already exists")
	ErrCallerIsNotParticipant = errors.New("caller is not participant")
)

type ErrUsersNotExists struct {
	Msg   string
	Users []uuid.UUID
}

func (err *ErrUsersNotExists) Error() string {
	return fmt.Sprintf("%s %v", err.Msg, err.Users)
}
