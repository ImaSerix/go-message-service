package query

import (
	"context"
	"time"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/google/uuid"
)

type ChatsQry interface {
	GetUserChats(ctx context.Context, userID uuid.UUID) ([]*GetUserChatsOutput, error)
}

type MessagesQry interface {
	GetUnreadCount(ctx context.Context, userID uuid.UUID, chatIDs []uuid.UUID) (map[uuid.UUID]int, error)
	GetMessages(ctx context.Context, cursorID *int64, chatID uuid.UUID, direction domain.Direction, limit int) ([]*domain.Message, bool, error)
	ExistsInChat(ctx context.Context, id int64, chatID uuid.UUID) (bool, error)
}

type GetUserChatsOutput struct {
	ID           uuid.UUID
	Name         string
	MessageCount int

	LastMessageID        *int64
	LastMessageSenderID  *uuid.UUID
	LastMessageContent   *string
	LastMessageCreatedAt *time.Time

	LastReadMessageID *int64

	CreatedAt time.Time
}
