package message

import (
	"time"

	"github.com/google/uuid"
)

type GetMessagesResponseMessage struct {
	ID        int64     `json:"id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type GetMessagesResponse struct {
	Messages     []GetMessagesResponseMessage `json:"messages"`
	CursorBefore *int64                       `json:"cursor_before"`
	CursorAfter  *int64                       `json:"cursor_after"`
}

type SendMessageRequest struct {
	Content string `json:"content"`
}

type SendMessageResponse struct {
	ID int64 `json:"id"`
}

type ReadMessageRequest struct {
	LastReadMessageID int64 `json:"last_read_message_id"`
}
