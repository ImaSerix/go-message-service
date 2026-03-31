package domain

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        int64
	ChatID    uuid.UUID
	SenderID  uuid.UUID
	Content   string
	CreatedAt time.Time
}
