package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type MessagesRepository struct {
	db *sql.DB
}

func (mr *MessagesRepository) GetUnreadCountForChats(ctx context.Context, userID uuid.UUID, chatIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	query := `SELECT 
				uc.chat_id AS c_id,
				COUNT(m.id) AS unread_count
			FROM users_chats uc
			LEFT JOIN messages m 
				ON m.chat_id = uc.chat_id
				AND m.id > COALESCE(uc.last_read_message_id, 0)
			WHERE uc.user_id = $1 AND chat_id = ANY ($2)
			GROUP BY uc.chat_id;`
	rows, err := mr.db.QueryContext(ctx, query, userID, chatIDs)
	if err != nil {
		return nil, fmt.Errorf("messages getUnreadCountForChats: %w", err)
	}
	defer rows.Close()

	unreadedCountByChat := make(map[uuid.UUID]int, len(chatIDs))
	for rows.Next() {
		var chatID uuid.UUID
		var count int
		if err := rows.Scan(&chatID, &count); err != nil {
			return nil, fmt.Errorf("messages getUnreadCountForChats: %w", err)
		}
		unreadedCountByChat[chatID] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("messages getUnreadCountForChats: %w", err)
	}
	return unreadedCountByChat, nil
}
func (mr *MessagesRepository) GetMessages(ctx context.Context, cursorID *int64, chatID uuid.UUID,
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

	rows, err := mr.db.QueryContext(ctx, query.String(), args...)
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

func (mr *MessagesRepository) Create(ctx context.Context, message *domain.Message) (*domain.Message, error) {

	tx, err := mr.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("message create: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `INSERT INTO messages (id, chat_id, sender_id, content) VALUES ($1, $2, $3, $4) RETURNING created_at`, message.ID, message.ChatID, message.SenderID, message.Content).Scan(&message.CreatedAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return nil, repository.ErrAlreadyExists
			}
		}
		return nil, fmt.Errorf("message create: %w", err)
	}

	_, err = tx.ExecContext(ctx, `UPDATE chats SET last_message_id = $1, last_message_sender_id = $2, last_message_count = $3, last_message_created_at = $4 WHERE id = $5`, message.ID, message.SenderID, message.Content, message.CreatedAt, message.ChatID)
	if err != nil {
		return nil, fmt.Errorf("message create: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("message create: %w", err)
	}

	return message, nil
}
func (mr *MessagesRepository) ExistsInChat(ctx context.Context, id int64, chatID uuid.UUID) (bool, error) {

	var messageExists int
	err := mr.db.QueryRowContext(ctx, `SELECT 1 FROM messages WHERE id = $1 and chat_id = $2`, id, chatID).Scan(&messageExists)
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
func (mr *MessagesRepository) UpdateLastReaded(ctx context.Context, userID uuid.UUID, messageID int64, chatID uuid.UUID) error {

	var updated int
	err := mr.db.QueryRowContext(ctx, `UPDATE users_chats SET last_read_message_id = $1 WHERE user_id = $2 and chat_id = $3 RETURNING 1`).Scan(&updated)
	if err != nil {
		return err
	}
	if updated != 1 {
		return fmt.Errorf("messages existsInChat: db returned something unexpected: %v", updated)
	}

	return nil
}
