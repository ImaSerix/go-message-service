package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/dto"
	"github.com/ImaSerix/go-message-service/internal/ws"
	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
)

type ChatsUseCase struct {
	chats    ChatsRepository
	messages MessagesRepository
	users    UsersRepository
	notifier Notifier
}

func New(c ChatsRepository, m MessagesRepository, u UsersRepository, n Notifier) *ChatsUseCase {
	return &ChatsUseCase{
		chats:    c,
		messages: m,
		users:    u,
		notifier: n,
	}
}

type CreateChatInput struct {
	Name         string
	Participants []uuid.UUID
}

func (cuc *ChatsUseCase) Create(ctx context.Context, in *CreateChatInput) (*dto.Chat, error) {
	existingIDs, err := cuc.users.GetExistingIDs(ctx, in.Participants)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	if len(existingIDs) != len(in.Participants) {
		diff := difference(in.Participants, existingIDs)
		return nil, &ErrUsersNotExists{
			msg:   "some participants not exists",
			users: diff,
		}
	}
	trimmedName := strings.TrimSpace(in.Name)
	if trimmedName == "" || len(trimmedName) < 2 || len(trimmedName) > 32 {
		return nil, ErrInvalidName
	}
	chat := &domain.Chat{
		ID:           uuid.New(),
		Name:         trimmedName,
		LastMessage:  nil,
		MessageCount: 0,
	}
	chat, err = cuc.chats.Create(ctx, chat, in.Participants)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}

	// TODO Спросить у чатика об этом говне
	cuc.notifier.NotifyChat(ctx, chat.ID, &wsevent.NewChat{
		Chat: ws.Chat{
			ID:           chat.ID,
			Name:         chat.Name,
			Participants: in.Participants,
		},
	})

	return &dto.Chat{
		ID:           chat.ID,
		Name:         chat.Name,
		LastMessage:  nil,
		UnreadCount:  0,
		MessageCount: 0,
	}, nil
}

type GetUserChatsInput struct {
	UserID uuid.UUID
}

func (cuc *ChatsUseCase) GetUserChats(ctx context.Context, in *GetUserChatsInput) ([]*dto.Chat, error) {
	chats, err := cuc.chats.GetUserChats(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("chat getByUser: %w", err)
	}

	chatsIDs := make([]uuid.UUID, 0, len(chats))
	for _, chat := range chats {
		chatsIDs = append(chatsIDs, chat.ID)
	}
	chatsParticipants, err := cuc.chats.GetChatsParticipants(ctx, chatsIDs)
	if err != nil {
		return nil, fmt.Errorf("chat getByUser: %w", err)
	}

	chatsUnreadMessagesCount, err := cuc.messages.GetUnreadCountForChats(ctx, in.UserID, chatsIDs)
	if err != nil {
		return nil, fmt.Errorf("chat getByUser: %w", err)
	}
	res := make([]*dto.Chat, 0, len(chats))
	for _, chat := range chats {
		var lastMessage *dto.Message
		if chat.LastMessage != nil {
			lastMessage = &dto.Message{
				ID:        chat.LastMessage.ID,
				ChatID:    chat.LastMessage.ChatID,
				SenderID:  chat.LastMessage.SenderID,
				Content:   chat.LastMessage.Content,
				CreatedAt: chat.LastMessage.CreatedAt,
			}
		}
		dtoChat := dto.Chat{
			ID:           chat.ID,
			Name:         chat.Name,
			Participants: chatsParticipants[chat.ID],
			LastMessage:  lastMessage,
			UnreadCount:  chatsUnreadMessagesCount[chat.ID],
			MessageCount: chat.MessageCount,
		}
		res = append(res, &dtoChat)
	}
	return res, nil
}
