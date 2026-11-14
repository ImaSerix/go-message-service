package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ImaSerix/go-message-service/internal/application/command"
	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/message"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type MessagesRepository struct {
	db *sql.DB
}

func NewMessagesRepository(db *sql.DB) *MessagesRepository {
	return &MessagesRepository{db: db}
}

func (r *MessagesRepository) Create(ctx context.Context, m *message.Message) (*message.Message, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("message create: %w", err)
	}
	defer tx.Rollback()

	var createdAt time.Time
	err = tx.QueryRowContext(ctx, `INSERT INTO messages (id, chat_id, sender_id, content) VALUES ($1, $2, $3, $4) RETURNING created_at`, int64(m.ID()), uuid.UUID(m.ChatID()), uuid.UUID(m.SenderID()), m.Content()).Scan(&createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, command.ErrAlreadyExists
		}
		return nil, fmt.Errorf("message create: %w", err)
	}
	m.SetCreatedAt(createdAt)

	if _, err := tx.ExecContext(ctx, `UPDATE chats SET last_message_id = $1, last_message_sender_id = $2, last_message_content = $3, last_message_created_at = $4, message_count = message_count + 1 WHERE id = $5`, int64(m.ID()), uuid.UUID(m.SenderID()), m.Content(), createdAt, uuid.UUID(m.ChatID())); err != nil {
		return nil, fmt.Errorf("message create: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("message create: %w", err)
	}
	return m, nil
}

func (r *MessagesRepository) ExistsInChat(ctx context.Context, messageID message.SnowflakeID, chatID chat.ChatID) (bool, error) {
	var messageExists int
	err := r.db.QueryRowContext(ctx, `SELECT 1 FROM messages WHERE id = $1 and chat_id = $2`, int64(messageID), uuid.UUID(chatID)).Scan(&messageExists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("messages repo existsInChat: %w", err)
	}

	return messageExists == 1, nil
}

func (r *MessagesRepository) UpdateLastReaded(ctx context.Context, readerID chat.ParticipantID, c *chat.Chat) error {
	lastReadedID, ok := c.LastReadMessageID(readerID)
	if !ok || lastReadedID == nil {
		return fmt.Errorf("messages updateLastReaded: no last read message")
	}

	var updated int
	err := r.db.QueryRowContext(ctx, `UPDATE chat_users SET last_read_message_id = $1 WHERE user_id = $2 and chat_id = $3 RETURNING 1`, int64(*lastReadedID), uuid.UUID(readerID), uuid.UUID(c.ID())).Scan(&updated)
	if err != nil {
		return fmt.Errorf("messages updateLastReaded: %w", err)
	}
	if updated != 1 {
		return fmt.Errorf("messages updateLastReaded: db returned something unexpected: %v", updated)
	}

	return nil
}
