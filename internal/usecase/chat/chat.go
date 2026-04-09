package chat

import (
	"context"
	"errors"
	"fmt"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/dto"
	"github.com/ImaSerix/go-message-service/internal/repository"
	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
)

type ChatsUseCase interface {
	Create(ctx context.Context, in *CreateChatInput) (*dto.Chat, error)
	GetUserChats(ctx context.Context, in *GetUserChatsInput) ([]*dto.Chat, error)
}

type chatsUseCase struct {
	chats    ChatsRepository
	messages MessagesRepository
	users    UsersRepository
	notifier Notifier
}

func New(c ChatsRepository, m MessagesRepository, u UsersRepository, n Notifier) ChatsUseCase {
	return &chatsUseCase{
		chats:    c,
		messages: m,
		users:    u,
		notifier: n,
	}
}

func (cuc *chatsUseCase) Create(ctx context.Context, in *CreateChatInput) (*dto.Chat, error) {
	trimmedName, err := in.Validate(ctx, cuc.users)
	if err != nil {
		return nil, err
	}

	existingIDs, err := cuc.users.GetExistingIDs(ctx, in.Participants)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	if len(existingIDs) != len(in.Participants) {
		diff := difference(in.Participants, existingIDs)
		return nil, &ErrUsersNotExists{
			Msg:   "some participants not exists",
			Users: diff,
		}
	}

	// Note: Репозиторий сохраняет новый чат, а также добавляет к нему пользователей-участников, так как чат без участников в пределах системы не имеет смысла
	chat := &domain.Chat{
		ID:           uuid.New(),
		Name:         trimmedName,
		LastMessage:  nil,
		MessageCount: 0,
	}
	chat, err = cuc.chats.Create(ctx, chat, in.Participants)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("create chat: %w", err)
	}

	go cuc.notifier.NotifyChat(ctx, chat.ID, &wsevent.NewChat{
		Chat: wsevent.Chat{
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

func (cuc *chatsUseCase) GetUserChats(ctx context.Context, in *GetUserChatsInput) ([]*dto.Chat, error) {

	chats, err := cuc.chats.GetUserChats(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("getUserChats: %w", err)
	}

	chatsIDs := make([]uuid.UUID, 0, len(chats))
	for _, chat := range chats {
		chatsIDs = append(chatsIDs, chat.ID)
	}
	chatsParticipants, err := cuc.chats.GetChatsParticipants(ctx, chatsIDs)
	if err != nil {
		return nil, fmt.Errorf("getUserChats: %w", err)
	}

	chatsUnreadMessagesCount, err := cuc.messages.GetUnreadCountForChats(ctx, in.UserID, chatsIDs)
	if err != nil {
		return nil, fmt.Errorf("getUserChats %w", err)
	}

	res := make([]*dto.Chat, 0, len(chats))
	for _, chat := range chats {
		var lastMessage *dto.Message
		if chat.LastMessage != nil {
			lastMessage = &dto.Message{
				ID:        int64(chat.LastMessage.ID),
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
