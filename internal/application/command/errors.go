package command

import (
	"errors"
	"fmt"

	"github.com/ImaSerix/go-message-service/internal/domain/user"
)

var (
	ErrAlreadyExists    = errors.New("already exists")
	ErrInvalidCommand   = errors.New("invalid command")
	ErrNotFound         = errors.New("invalid command")
	ErrChatNotExists    = errors.New("chat not exists")
	ErrMessageNotInChat = errors.New("messsage not in chat")
	UserAlreadyExists   = errors.New("user already exists")
)

type ErrUsersNotExist struct {
	Msg   string
	Users []user.UserID
}

func (err *ErrUsersNotExist) Error() string {
	return fmt.Sprintf("%s %v", err.Msg, err.Users)
}
