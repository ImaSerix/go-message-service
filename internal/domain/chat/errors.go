package chat

import "errors"

var (
	ErrInvalidChatName        = errors.New("invalid chat name")
	ErrCallerIsNotParticipant = errors.New("caller is not participant")
	ErrParticipantsEmpty      = errors.New("participants empty")
)
