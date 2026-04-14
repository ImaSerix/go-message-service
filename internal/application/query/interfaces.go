package query

import (
	"context"
	"time"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/google/uuid"
)

type ChatsQry interface {
	GetUserChat(ctx context.Context, userID uuid.UUID, chatID uuid.UUID) (*GetUserChatOutput, error)
	GetUserChats(ctx context.Context, userID uuid.UUID) ([]*GetUserChatsOutput, error)
}

type MessagesQry interface {
	GetUnreadCount(ctx context.Context, userID uuid.UUID, chatIDs []uuid.UUID) (map[uuid.UUID]int, error)
	GetMessages(ctx context.Context, cursorID *int64, chatID uuid.UUID, direction domain.Direction, limit int) ([]*domain.Message, bool, error)
	ExistsInChat(ctx context.Context, id int64, chatID uuid.UUID) (bool, error)
}

type GetUserChatOutput struct {
	ID           uuid.UUID
	Name         string
	MessageCount int

	Participants []uuid.UUID

	LastMessageID        *int64
	LastMessageSenderID  *uuid.UUID
	LastMessageContent   *string
	LastMessageCreatedAt *time.Time

	LastReadMessageID *int64

	CreatedAt time.Time
}

// TODO: подумать об этом тоже, как будто да, но я не знаю
// NOTE: фактически, если запрашивать просто чаты, то наверное не нужно особо информация о кол-во сообщений в чате, а также о участниках
type GetUserChatsOutput struct {
	ID   uuid.UUID
	Name string

	LastMessageID        *int64
	LastMessageSenderID  *uuid.UUID
	LastMessageContent   *string
	LastMessageCreatedAt *time.Time

	LastReadMessageID *int64

	CreatedAt time.Time
}
