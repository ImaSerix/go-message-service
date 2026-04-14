package query

import (
	"context"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/google/uuid"
)

type ChatsQry interface {
	GetChat(ctx context.Context, chatID uuid.UUID) (*GetUserOutput, error)
	GetUserChats(ctx context.Context, userID uuid.UUID) ([]*GetUserChatsOutput, error)
}

type MessagesQry interface {
	GetUnreadCount(ctx context.Context, userID uuid.UUID, chatIDs []uuid.UUID) (map[uuid.UUID]int, error)
	GetMessages(ctx context.Context, cursorID *int64, chatID uuid.UUID, direction domain.Direction, limit int) ([]*domain.Message, bool, error)
	ExistsInChat(ctx context.Context, id int64, chatID uuid.UUID) (bool, error)
}
