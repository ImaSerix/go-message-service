package message

import (
	"context"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
)

type MessagesRepository interface {
	GetMessages(ctx context.Context, cursorID *int64, chatID uuid.UUID, direction domain.Direction, limit int) ([]*domain.Message, bool, error)
	Create(ctx context.Context, message *domain.Message) (*domain.Message, error)
	ExistsInChat(ctx context.Context, id int64, chatID uuid.UUID) (bool, error)
	UpdateLastReaded(ctx context.Context, userID uuid.UUID, messageID int64, chatID uuid.UUID) error
	UpdateLastMessage //TODO Сделать вот это
}

type ChatsRepository interface {
	UserIsParticipant(ctx context.Context, userID uuid.UUID, id uuid.UUID) (bool, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type Notifier interface {
	NotifyChat(ctx context.Context, chatID uuid.UUID, event wsevent.WSEvent)
}

type IDGenerator interface {
	NewID() (int64, error)
}
