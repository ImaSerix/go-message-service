package chat

import (
	"context"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
)

type ChatsRepository interface {
	Create(ctx context.Context, chat *domain.Chat, participants []uuid.UUID) (*domain.Chat, error) // Makes chat and adds users as participants
	GetUserChats(ctx context.Context, userID uuid.UUID) ([]*domain.Chat, error)
	GetChatsParticipants(ctx context.Context, chatIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error)
}

type UsersRepository interface {
	GetExistingIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error)
}

type MessagesRepository interface {
	GetUnreadCountForChats(ctx context.Context, userID uuid.UUID, chatIDs []uuid.UUID) (map[uuid.UUID]int, error)
}

// TODO Нужно добавить это в обсидиан заметки
type Notifier interface {
	NotifyChat(ctx context.Context, chatID uuid.UUID, event wsevent.WSEvent)
}
