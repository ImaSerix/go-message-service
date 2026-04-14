package query

import (
	"context"
	"fmt"

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

func (qs *QueryService) HandleGetChat(ctx context.Context, qry *GetUserChatQry) (*dto.Chat, error) {
	// TODO: это тоже надо реапизовать
}

// TODO: Наверное стоит немного урезать возвращаемый чатами контент
func (qs *QueryService) HandleGetUserChats(ctx context.Context, qry *GetUserChatsQry) ([]*dto.Chat, error) {
	chats, err := qs.chats.GetUserChats(ctx, qry.UserID)
	if err != nil {
		return nil, err
	}

	chatIDs := make([]uuid.UUID, 0, len(chats))
	for _, c := range chats {
		chatIDs = append(chatIDs, c.ID)
	}

	unreadCountByChat, err := qs.messages.GetUnreadCount(ctx, qry.UserID, chatIDs)
	if err != nil {
		return nil, err
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
		Messages:  toDTOList(messages),
		HasMore:   hasMore,
		NewCursor: &newCursor,
	}, nil
}
