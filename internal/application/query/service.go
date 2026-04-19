package query

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/ImaSerix/go-message-service/internal/dto"
	"github.com/google/uuid"
)

type QueryService struct {
	chats    ChatsQry
	messages MessagesQry
	limit    int
}

func NewQueryService(chats ChatsQry, messages MessagesQry, limit int) *QueryService {
	return &QueryService{
		chats:    chats,
		messages: messages,
		limit:    limit,
	}
}

// NOTE: Фактически запрос на то, чтоб вернуть участников и кол-во сообщений, не знаю зачем этот запрос может быть нужен, но пусть будет
type GetUserOutput struct {
	ID   uuid.UUID
	Name string

	MessageCount int
	Participants []uuid.UUID

	CreatedAt time.Time
}

func (qs *QueryService) HandleGetUserChat(ctx context.Context, qry *GetUserChatQry) (*dto.Chat, error) {
	c, err := qs.chats.GetChat(ctx, qry.ChatID)
	if err != nil {
		return nil, fmt.Errorf("handleGetUserChat: %w", err)
	}

	if !slices.Contains(c.Participants, qry.UserID) {
		return nil, ErrUserIsNotParticipant
	}

	return &dto.Chat{
		ID:           c.ID,
		Name:         c.Name,
		MessageCount: c.MessageCount,
		Participants: c.Participants,
		CreatedAt:    c.CreatedAt,
	}, nil
}

// NOTE: фактически, если запрашивать просто чаты, то наверное не нужно особо информация о кол-во сообщений в чате, а также о участниках
type GetUserChatsOutput struct {
	ID   uuid.UUID
	Name string

	LastMessageID        *int64
	LastMessageSenderID  *uuid.UUID
	LastMessageContent   *string
	LastMessageCreatedAt *time.Time

	LastReadMessageID *int64

	CreatedAt time.Time
}

func (qs *QueryService) HandleGetUserChats(ctx context.Context, qry *GetUserChatsQry) ([]*dto.Chat, error) {
	chats, err := qs.chats.GetUserChats(ctx, qry.UserID)
	if err != nil {
		return nil, fmt.Errorf("handleGetUserChats: %w", err)
	}

	chatIDs := make([]uuid.UUID, 0, len(chats))
	for _, c := range chats {
		chatIDs = append(chatIDs, c.ID)
	}

	unreadCountByChat, err := qs.messages.GetUnreadCount(ctx, qry.UserID, chatIDs)
	if err != nil {
		return nil, fmt.Errorf("handleGetUserChats: %w", err)
	}

	res := make([]*dto.Chat, 0, len(chats))
	for _, chat := range chats {
		dtoChat := dto.Chat{
			ID:          chat.ID,
			Name:        chat.Name,
			UnreadCount: unreadCountByChat[chat.ID],

			LastMessageID:        chat.LastMessageID,
			LastMessageSenderID:  chat.LastMessageSenderID,
			LastMessageContent:   chat.LastMessageContent,
			LastMessageCreatedAt: chat.LastMessageCreatedAt,

			LastReadMessageID: chat.LastReadMessageID,

			CreatedAt: chat.CreatedAt,
		}
		res = append(res, &dtoChat)
	}
	return res, nil
}

type GetMessageOutput struct {
	Messages  []*dto.Message
	HasMore   bool
	NewCursor *int64
}

func (qs *QueryService) HandleGetUserMessages(ctx context.Context, qry *GetUserMessagesQry) (*GetMessageOutput, error) {

	if qry.Cursor != nil {
		exists, err := qs.messages.ExistsInChat(ctx, *qry.Cursor, qry.ChatID)
		if err != nil {
			return nil, fmt.Errorf("get messages: %w", err)
		}
		if !exists {
			return nil, ErrMessageNotInChat
		}
	}

	if qry.Limit > qs.limit {
		return nil, ErrInvalidLimit
	}

	messages, hasMore, err := qs.messages.GetMessages(ctx, qry.Cursor, qry.ChatID, qry.Direction, qry.Limit)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}
	if len(messages) == 0 {
		return &GetMessageOutput{
			Messages:  []*dto.Message{},
			HasMore:   hasMore,
			NewCursor: nil,
		}, nil
	}

	newCursor := int64(messages[len(messages)-1].ID)

	return &GetMessageOutput{
		Messages:  messages,
		HasMore:   hasMore,
		NewCursor: &newCursor,
	}, nil
}
