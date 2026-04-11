package domain

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        SnowflakeID
	ChatID    uuid.UUID
	SenderID  uuid.UUID
	Content   string
	CreatedAt time.Time
}
