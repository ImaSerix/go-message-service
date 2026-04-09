package message

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

type MessagesUseCase interface {
	Get(ctx context.Context, in *GetMessagesInput) (*GetMessageOutput, error)
	Create(ctx context.Context, in *CreateMessageInput) (*dto.Message, error)
	MarkAsRead(ctx context.Context, in *ReadMessageInput) error
}

type messagesUseCase struct {
	messagеs    MessagesRepository
	chats       ChatsRepository
	notifier    Notifier
	idGenerator IDGenerator
	limit       int
}

func New(m MessagesRepository, limit int, c ChatsRepository, n Notifier, g IDGenerator) MessagesUseCase {
	return &messagesUseCase{
		messagеs:    m,
		chats:       c,
		notifier:    n,
		idGenerator: g,
		limit:       limit,
	}
}

func (uuc *messagesUseCase) Get(ctx context.Context, in *GetMessagesInput) (*GetMessageOutput, error) {

	isParticipant, err := uuc.chats.UserIsParticipant(ctx, in.UserID, in.ChatID)
	if err != nil {
		return nil, fmt.Errorf("get message: %w", err)
	}
	if !isParticipant {
		return nil, ErrUserIsNotParticipant
	}

	exists, err := uuc.chats.Exists(ctx, in.ChatID)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}
	if !exists {
		return nil, ErrChatNotExists
	}

	if in.Cursor != nil {
		exists, err := uuc.messagеs.ExistsInChat(ctx, *in.Cursor, in.ChatID)
		if err != nil {
			return nil, fmt.Errorf("get messages: %w", err)
		}
		if !exists {
			return nil, ErrMessageNotInChat
		}
	}
	// TODO: Нужно ограничить максимальный лимит, ибо нефиг

	if in.Limit > uuc.limit {
		return nil, ErrInvalidLimit
	}

	messages, hasMore, err := uuc.messagеs.GetMessages(ctx, in.Cursor, in.ChatID, in.Direction, in.Limit)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	newCursor := int64(messages[len(messages)-1].ID)

	return &GetMessageOutput{
		Messages:  toDTOList(messages),
		HasMore:   hasMore,
		NewCursor: &newCursor,
	}, nil
}

func (uuc *messagesUseCase) Create(ctx context.Context, in *CreateMessageInput) (*dto.Message, error) {

	if err := in.Validate(); err != nil {
		return nil, err
	}

	isParticipant, err := uuc.chats.UserIsParticipant(ctx, in.UserID, in.ChatID)
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}
	if !isParticipant {
		return nil, ErrUserIsNotParticipant
	}

	exists, err := uuc.chats.Exists(ctx, in.ChatID)
	if err != nil {
		return nil, fmt.Errorf("create messages: %w", err)
	}
	if !exists {
		return nil, ErrChatNotExists
	}

	id := uuc.idGenerator.NewID()

	// NOTE: Создавая сообщение также обновляется информация о последнем сообщении в чате
	message := &domain.Message{
		ID:       id,
		ChatID:   in.ChatID,
		SenderID: in.UserID,
		Content:  in.Content,
	}
	message, err = uuc.messagеs.Create(ctx, message)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("create messages: %w", err)
	}

	uuc.notifier.NotifyChat(ctx, in.ChatID, &wsevent.NewMessage{
		Message: wsevent.Message{
			ID:        int64(message.ID),
			ChatID:    message.ChatID,
			SenderID:  message.SenderID,
			Content:   message.Content,
			CreatedAt: message.CreatedAt,
		},
	})
	return toDTO(message), nil
}

type ReadMessageInput struct {
	ChatID    uuid.UUID
	UserID    uuid.UUID
	MessageID int64
}

func (uuc *messagesUseCase) MarkAsRead(ctx context.Context, in *ReadMessageInput) error {

	isParticipant, err := uuc.chats.UserIsParticipant(ctx, in.UserID, in.ChatID)
	if err != nil {
		return fmt.Errorf("markAsRead message: %w", err)
	}
	if !isParticipant {
		return ErrUserIsNotParticipant
	}

	exists, err := uuc.chats.Exists(ctx, in.ChatID)
	if err != nil {
		return fmt.Errorf("markAsRead messages: %w", err)
	}
	if !exists {
		return ErrChatNotExists
	}

	exists, err = uuc.messagеs.ExistsInChat(ctx, in.MessageID, in.ChatID)
	if err != nil {
		return fmt.Errorf("markAsRead messages: %w", err)
	}
	if !exists {
		return ErrMessageNotInChat
	}

	err = uuc.messagеs.UpdateLastReaded(ctx, in.UserID, in.MessageID, in.ChatID)
	if err != nil {
		return fmt.Errorf("markAsRead messages: %w", err)
	}
	return nil
}
