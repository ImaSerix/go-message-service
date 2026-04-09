package notifier

import (
	"context"

	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
)

type ChatsRepository interface {
	GetChatParticipants(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error)
}

type WsHUB interface {
	SendToUsers(ids []uuid.UUID, event wsevent.WSEvent)
}
