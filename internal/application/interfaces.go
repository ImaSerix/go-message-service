package application

import (
	"context"

	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/ImaSerix/go-message-service/internal/wsevent"
)

type ChatsRepository interface {
	Create(ctx context.Context, chat *chat.Chat) (*chat.Chat, error)
}

type UsersRepository interface {
	GetExistingIDs(ctx context.Context, ids []user.UserID) ([]user.UserID, error)
}

type MessagesRepository interface {
}

type Notifier interface {
	NotifyChat(ctx context.Context, id chat.ChatID, event wsevent.WSEvent)
}
