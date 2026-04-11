package chat

import "errors"

var (
	ErrInvalidChatName           = errors.New("invalid chat name")
	ErrCallerIsNotParticipant    = errors.New("caller is not participant")
	ErrSenderIsNotParticipant    = errors.New("sender is not participant")
	ErrParticipantsEmpty         = errors.New("participants empty")
	ErrMessageNotInChat          = errors.New("message not in chat")
	ErrReaderIsNotParticipant    = errors.New("reader is not participant")
	ErrNewerMessageIsAlreadyRead = errors.New("newer message is already read")
)
