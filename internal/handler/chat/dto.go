package chat

import (
	"time"

	"github.com/google/uuid"
)

type GetUserResponseLastMessage struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	SenderID  uuid.UUID `json:"sender_id"`
	CreatedAt time.Time `json:"created_at"`
}

type GetUserResponseChat struct {
	ID           uuid.UUID                   `json:"id"`
	Name         string                      `json:"name"`
	Participants []uuid.UUID                 `json:"participants"`
	LastMessage  *GetUserResponseLastMessage `json:"last_message"`
	MessageCount int                         `json:"message_count"`
	UnreadCount  int                         `json:"unread_count"`
}

type GetUserResponse struct {
	Chats []GetUserResponseChat `json:"chats"`
}

type CreateChatRequest struct {
	Name         string      `json:"name"`
	Participants []uuid.UUID `json:"participants"`
}

type CreateChatResponse struct {
	ID uuid.UUID `json:"id"`
}
