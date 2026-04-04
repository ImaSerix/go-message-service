package message

import (
	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/google/uuid"
)

type GetMessagesInput struct {
	ChatID    uuid.UUID
	UserID    uuid.UUID
	Cursor    *int64
	Direction domain.Direction
}

type CreateMessageInput struct {
	ChatID  uuid.UUID
	UserID  uuid.UUID
	Content string
}

func (in *CreateMessageInput) Validate() error {
	if len(in.Content) == 0 {
		return ErrContentEmpty
	}
	return nil
}
