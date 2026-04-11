package query

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type QueryService struct {
	db    *sql.DB
	limit int
}

// TODO: Наверное стоит немного урезать возвращаемый чатами контент
func (qs *QueryService) HandleGetUserChats(ctx context.Context, qry *GetUserChatsQry) ([]*dto.Chat, error) {
	chats, err := qs.getUserChats(ctx, qry.UserID)
	if err != nil {
		return nil, err
	}

	chatIDs := make([]uuid.UUID, 0, len(chats))
	for _, c := range chats {
		chatIDs = append(chatIDs, c.id)
	}

	unreadCountByChat, err := qs.getUnreadCount(ctx, qry.UserID, chatIDs)
	if err != nil {
		return nil, err
	}

	res := make([]*dto.Chat, 0, len(chats))
	for _, chat := range chats {
		dtoChat := dto.Chat{
			ID:           chat.id,
			Name:         chat.name,
			MessageCount: chat.messageCount,
			UnreadCount:  unreadCountByChat[chat.id],

			LastMessageID:        chat.lastMessageID,
			LastMessageSenderID:  chat.lastMessageSenderID,
			LastMessageContent:   chat.lastMessageContent,
			LastMessageCreatedAt: chat.lastMessageCreatedAt,

			LastReadMessageID: chat.lastReadMessageID,

			CreatedAt: chat.createdAt,
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
		exists, err := qs.existsInChat(ctx, *qry.Cursor, qry.ChatID)
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

	messages, hasMore, err := qs.getMessages(ctx, qry.Cursor, qry.ChatID, qry.Direction, qry.Limit)
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

type getUserChatsOutput struct {
	id           uuid.UUID
	name         string
	messageCount int

	lastMessageID        *int64
	lastMessageSenderID  *uuid.UUID
	lastMessageContent   *string
	lastMessageCreatedAt *time.Time

	lastReadMessageID *int64

	createdAt time.Time
}

func (qs *QueryService) getUserChats(ctx context.Context, userID uuid.UUID) ([]*getUserChatsOutput, error) {

	query := `SELECT 
					c.id, 
					c.name, 
					c.message_count, 
					c.last_message_id, 
					c.last_message_sender_id, 
					c.last_message_content, 
					c.last_message_created_at,
					cu.last_read_message_id,
					c.createdAt
				FROM chat_users cu
				INNER JOIN chats c
				ON cu.chat_id = c.id
				WHERE cu.user_id = $1;`
	rows, err := qs.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("getUserChats: %w", err)
	}
	defer rows.Close()

	res := []*getUserChatsOutput{}
	for rows.Next() {
		c := getUserChatsOutput{}
		if err := rows.Scan(
			c.id,
			c.name,
			c.messageCount,
			c.lastMessageID,
			c.lastMessageSenderID,
			c.lastMessageContent,
			c.lastMessageCreatedAt,
			c.lastReadMessageID,
			c.createdAt,
		); err != nil {
			return nil, fmt.Errorf("getUserChats: %w", err)
		}
		res = append(res, &c)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("getUserChats: %w", err)
	}
	return res, nil
}

func (qs *QueryService) getUnreadCount(ctx context.Context, userID uuid.UUID, chatIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	query := `SELECT 
				uc.chat_id AS c_id,
				COUNT(m.id) AS unread_count
			FROM users_chats uc
			LEFT JOIN messages m 
			ON m.chat_id = uc.chat_id
			AND m.id > COALESCE(uc.last_read_message_id, 0)
			WHERE uc.user_id = $1 AND uc.chat_id = ANY ($2)
			GROUP BY uc.chat_id;`

	rows, err := qs.db.QueryContext(ctx, query, userID, chatIDs)
	if err != nil {
		return nil, fmt.Errorf("getUnreadCount: %w", err)
	}
	defer rows.Close()

	unreadedCountByChat := make(map[uuid.UUID]int, len(chatIDs))
	for rows.Next() {
		var chatID uuid.UUID
		var count int
		if err := rows.Scan(&chatID, &count); err != nil {
			return nil, fmt.Errorf("getUnreadCount: %w", err)
		}
		unreadedCountByChat[chatID] = count
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("getUnreadCount: %w", err)
	}
	return unreadedCountByChat, rows.Err()
}

func (qs *QueryService) existsInChat(ctx context.Context, id int64, chatID uuid.UUID) (bool, error) {

	var messageExists int
	err := qs.db.QueryRowContext(ctx, `SELECT 1 FROM messages WHERE id = $1 and chat_id = $2`, id, chatID).Scan(&messageExists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("messages existsInChat: %w", err)
	}
	if messageExists != 1 {
		return false, fmt.Errorf("messages existsInChat: db returned something unexpected: %v", messageExists)
	}

	return true, nil
}

func (qs *QueryService) getMessages(ctx context.Context, cursorID *int64, chatID uuid.UUID,
	direction domain.Direction, limit int) ([]*domain.Message, bool, error) {
	var query strings.Builder

	query.WriteString("SELECT id, chat_id, sender_id, content, created_at FROM messages WHERE chat_id = $1")

	args := []any{chatID}
	args = append(args, chatID)
	if cursorID != nil {
		if direction == domain.DirectionForward {
			query.WriteString(" AND id > $2 ORDER BY id ASC")
		} else {
			query.WriteString(" AND id < $2 ORDER BY id DESC")
		}
		args = append(args, *cursorID)
	} else {
		if direction == domain.DirectionForward {
			query.WriteString(" ORDER BY id ASC")
		} else {
			query.WriteString(" ORDER BY id DESC")
		}
	}

	query.WriteString(fmt.Sprintf(" LIMIT %d", limit+1))

	rows, err := qs.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, false, fmt.Errorf("messages get: %w", err)
	}
	defer rows.Close()

	messages := make([]*domain.Message, 0, limit+1)
	for rows.Next() {
		var message domain.Message
		if err := rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.SenderID,
			&message.Content,
			&message.CreatedAt,
		); err != nil {
			return nil, false, fmt.Errorf("messages get: %w", err)
		}
		messages = append(messages, &message)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("messages get: %w", err)
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	if direction == domain.DirectionBackward {
		slices.Reverse(messages)
	}

	return messages, hasMore, nil
}
