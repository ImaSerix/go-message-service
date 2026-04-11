package dto

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID           uuid.UUID
	Name         string
	MessageCount int
	UnreadCount  int

	LastMessageID        *int64
	LastMessageSenderID  *uuid.UUID
	LastMessageContent   *string
	LastMessageCreatedAt *time.Time

	LastReadMessageID *int64

	CreatedAt time.Time
}
