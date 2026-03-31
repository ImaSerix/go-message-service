package message

import (
	"context"
	"fmt"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/dto"
	"github.com/ImaSerix/go-message-service/internal/ws"
	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
)

type MessageUseCase struct {
	messagеs    MessagesRepository
	chats       ChatsRepository
	users       UsersRepository
	notifier    Notifier
	idGenerator IDGenerator
	limit       int
}

func New(m MessagesRepository, limit int, c ChatsRepository, u UsersRepository, n Notifier, g IDGenerator) *MessageUseCase {
	return &MessageUseCase{
		messagеs:    m,
		chats:       c,
		users:       u,
		notifier:    n,
		idGenerator: g,
		limit:       limit,
	}
}

type Direction string

const (
	DirectionForward  Direction = "forward"  // newer than cursor
	DirectionBackward Direction = "backward" // older than cursor
)

type GetMessagesInput struct {
	ChatID    uuid.UUID
	UserID    uuid.UUID
	Cursor    *int64
	Direction Direction
}

type GetMessageOutput struct {
	Messages  []*dto.Message
	HasMore   bool
	NewCursor *int64
}

func (uuc *MessageUseCase) Get(ctx context.Context, in *GetMessagesInput) (*GetMessageOutput, error) {
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

	messages, hasMore, err := uuc.messagеs.GetMessages(ctx, in.Cursor, in.ChatID, in.Direction, 50)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}
	return &GetMessageOutput{
		Messages:  toDTOList(messages),
		HasMore:   hasMore,
		NewCursor: &messages[len(messages)-1].ID,
	}, nil
}

type CreateMessageInput struct {
	ChatID  uuid.UUID
	UserID  uuid.UUID
	Content string
}

func (uuc *MessageUseCase) Create(ctx context.Context, in *CreateMessageInput) (*dto.Message, error) {
	if len(in.Content) == 0 {
		return nil, ErrContentEmpty
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
	id, err := uuc.idGenerator.NewID()
	if err != nil {
		return nil, fmt.Errorf("create messages: failed generate id: %w", err)
	}

	message := &domain.Message{
		ID:       id,
		ChatID:   in.ChatID,
		SenderID: in.UserID,
		Content:  in.Content,
	}
	message, err = uuc.messagеs.Create(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("create messages: %w", err)
	}
	uuc.notifier.NotifyChat(ctx, in.ChatID, &wsevent.NewMessage{
		Message: ws.Message{
			ID:        message.ID,
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

func (uuc *MessageUseCase) MarkAsRead(ctx context.Context, in *ReadMessageInput) error {
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
	err = uuc.messagеs.UpdateLastReaded(ctx, in.MessageID, in.ChatID)
	if err != nil {
		return fmt.Errorf("markAsRead messages: %w", err)
	}
	return nil
}
