package wsevent

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        int64     `json:"id"`
	ChatID    uuid.UUID `json:"chat_id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Chat struct {
	ID           uuid.UUID   `json:"id"`
	Name         string      `json:"name"`
	Participants []uuid.UUID `json:"participants"`
}
