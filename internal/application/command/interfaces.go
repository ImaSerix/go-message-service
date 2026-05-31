package command

import (
	"context"

	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/message"
	"github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
)

type ChatsRepository interface {
	Create(ctx context.Context, chat *chat.Chat) (*chat.Chat, error)
	LoadForSend(ctx context.Context, id chat.ChatID) (*chat.Chat, error)
	LoadForRead(ctx context.Context, id chat.ChatID) (*chat.Chat, error)
}

type UsersRepository interface {
	GetExistingIDs(ctx context.Context, ids []user.UserID) ([]user.UserID, error)
	Create(ctx context.Context, user *user.User) (*user.User, error)
}

type MessagesRepository interface {
	Create(ctx context.Context, message *message.Message) (*message.Message, error)
	ExistsInChat(ctx context.Context, messageID message.SnowflakeID, chatID chat.ChatID) (bool, error)
	UpdateLastReaded(ctx context.Context, readerID chat.ParticipantID, chat *chat.Chat) error
}

type Notifier interface {
	SendToUsers(participants []uuid.UUID, event wsevent.WSEvent)
}

type IDGenerator interface {
	NewID() message.SnowflakeID
}

type Runner interface {
	Run(func())
}
