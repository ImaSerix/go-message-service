package notifier

import (
	"context"
	"log/slog"

	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
)

type Notifier struct {
	chats ChatsRepository
	hub   WsHUB
}

func New(chats ChatsRepository, hub WsHUB) *Notifier {
	return &Notifier{
		chats: chats,
		hub:   hub,
	}
}

func (n *Notifier) NotifyChat(ctx context.Context, chatID uuid.UUID, event wsevent.WSEvent) {

	participants, err := n.chats.GetChatParticipants(ctx, chatID)
	if err != nil {
		slog.Error("failed get chat participants", "err", err, "chatID", chatID, "eventName", event.Name())
		return
	}

	n.hub.SendToUsers(participants, event)
}
