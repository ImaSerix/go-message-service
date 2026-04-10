package wsevent

import (
	"time"

	"github.com/ImaSerix/go-message-service/internal/domain/chat"
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
	ID           chat.ChatID          `json:"id"`
	Name         chat.Name            `json:"name"`
	Participants []chat.ParticipantID `json:"participants"`
}
