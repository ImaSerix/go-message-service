package wsevent

import (
	"time"

	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/message"
)

type Message struct {
	ID        message.SnowflakeID `json:"id"`
	ChatID    message.ChatID      `json:"chat_id"`
	SenderID  message.SenderID    `json:"sender_id"`
	Content   message.Content     `json:"content"`
	CreatedAt time.Time           `json:"created_at"`
}

type Chat struct {
	ID           chat.ChatID          `json:"id"`
	Name         chat.Name            `json:"name"`
	Participants []chat.ParticipantID `json:"participants"`
}
