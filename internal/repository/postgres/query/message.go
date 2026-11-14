package query

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/dto"
	"github.com/google/uuid"
)

type MessageQry struct {
	db *sql.DB
}

func NewMessageQry(db *sql.DB) *MessageQry {
	return &MessageQry{
		db: db,
	}
}

func (q *MessageQry) GetUnreadCount(ctx context.Context, userID uuid.UUID, chatIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	query := `SELECT 
				uc.chat_id AS c_id,
				COUNT(m.id) AS unread_count
			FROM chat_users uc
			LEFT JOIN messages m 
			ON m.chat_id = uc.chat_id
			AND m.id > COALESCE(uc.last_read_message_id, 0)
			WHERE uc.user_id = $1 AND uc.chat_id = ANY ($2)
			GROUP BY uc.chat_id;`

	rows, err := q.db.QueryContext(ctx, query, userID, chatIDs)
	if err != nil {
		return nil, fmt.Errorf("messages getUnreadCount: %w", err)
	}
	defer rows.Close()

	unreadedCountByChat := make(map[uuid.UUID]int, len(chatIDs))
	for rows.Next() {
		var chatID uuid.UUID
		var count int
		if err := rows.Scan(&chatID, &count); err != nil {
			return nil, fmt.Errorf("messages getUnreadCount: %w", err)
		}
		unreadedCountByChat[chatID] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("messages getUnreadCount: %w", err)
	}
	return unreadedCountByChat, nil
}

func (q *MessageQry) GetMessages(ctx context.Context, cursorID *int64, chatID uuid.UUID,
	direction domain.Direction, limit int) ([]*dto.Message, bool, error) {
	var query strings.Builder

	query.WriteString("SELECT id, chat_id, sender_id, content, created_at FROM messages WHERE chat_id = $1")

	args := []any{chatID}
	if cursorID != nil {
		if direction == domain.DirectionForward {
			query.WriteString(" AND id > $2 ORDER BY id ASC")
		} else {
			query.WriteString(" AND id < $2 ORDER BY id DESC")
		}
		args = append(args, *cursorID)
	} else {
		query.WriteString(" ORDER BY id DESC")
	}

	query.WriteString(fmt.Sprintf(" LIMIT %d", limit+1))

	rows, err := q.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, false, fmt.Errorf("messages get: %w", err)
	}
	defer rows.Close()

	messages := make([]*dto.Message, 0, limit+1)
	for rows.Next() {
		var message dto.Message
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

func (q *MessageQry) ExistsInChat(ctx context.Context, messageID int64, chatID uuid.UUID) (bool, error) {
	var messageExists int
	err := q.db.QueryRowContext(ctx, `SELECT 1 FROM messages WHERE id = $1 and chat_id = $2`, messageID, chatID).Scan(&messageExists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("messages qry existsInChat: %w", err)
	}

	return messageExists == 1, nil
}
