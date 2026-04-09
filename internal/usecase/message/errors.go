package message

import "errors"

var (
	ErrUserIsNotParticipant = errors.New("user is not participant")
	ErrChatNotExists        = errors.New("chat not exists")
	ErrMessageNotInChat     = errors.New("message not in chat")
	ErrContentEmpty         = errors.New("content empty")
	ErrAlreadyExists        = errors.New("already exists")
	ErrInvalidLimit         = errors.New("invalid limit")
)
