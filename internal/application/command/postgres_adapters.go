package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/message"
	"github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PGChatsRepository struct {
	db *sql.DB
}

func NewPGChatsRepository(db *sql.DB) *PGChatsRepository {
	return &PGChatsRepository{db: db}
}

func (r *PGChatsRepository) Create(ctx context.Context, c *chat.Chat) (*chat.Chat, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `INSERT INTO chats (id, name) VALUES ($1, $2)`, c.ID(), c.Name())
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("create chat: %w", err)
	}

	participants := idSliceConverter[chat.ParticipantID, [16]byte](c.Participants())
	if _, err := tx.ExecContext(ctx, `INSERT INTO chat_users (chat_id, user_id) SELECT $1, unnest($2::uuid[])`, c.ID(), participants); err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	return c, nil
}

func (r *PGChatsRepository) LoadForSend(ctx context.Context, id chat.ChatID) (*chat.Chat, error) {
	return r.loadChat(ctx, id)
}

func (r *PGChatsRepository) LoadForRead(ctx context.Context, id chat.ChatID) (*chat.Chat, error) {
	return r.loadChat(ctx, id)
}

func (r *PGChatsRepository) loadChat(ctx context.Context, id chat.ChatID) (*chat.Chat, error) {
	var name string
	err := r.db.QueryRowContext(ctx, `SELECT name FROM chats WHERE id = $1`, id).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("load chat: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `SELECT user_id FROM chat_users WHERE chat_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("load chat: %w", err)
	}
	defer rows.Close()

	participants := make([]chat.ParticipantID, 0)
	for rows.Next() {
		var participant [16]byte
		if err := rows.Scan(&participant); err != nil {
			return nil, fmt.Errorf("load chat: %w", err)
		}
		participants = append(participants, idConverter[[16]byte, chat.ParticipantID](participant))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load chat: %w", err)
	}
	if len(participants) == 0 {
		return nil, fmt.Errorf("load chat: chat has no participants")
	}

	chatName, err := chat.NewName(name)
	if err != nil {
		return nil, fmt.Errorf("load chat: %w", err)
	}

	return chat.NewChat(id, chatName, participants[0], participants)
}

type PGUsersRepository struct {
	db *sql.DB
}

func NewPGUsersRepository(db *sql.DB) *PGUsersRepository {
	return &PGUsersRepository{db: db}
}

func (r *PGUsersRepository) GetExistingIDs(ctx context.Context, ids []user.UserID) ([]user.UserID, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id FROM users WHERE id = ANY($1::uuid[])`, idSliceConverter[user.UserID, [16]byte](ids))
	if err != nil {
		return nil, fmt.Errorf("users GetExistingIDs: %w", err)
	}
	defer rows.Close()

	result := make([]user.UserID, 0)
	for rows.Next() {
		var id [16]byte
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("users GetExistingIDs: %w", err)
		}
		result = append(result, idConverter[[16]byte, user.UserID](id))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("users GetExistingIDs: %w", err)
	}

	return result, nil
}

func (r *PGUsersRepository) Create(ctx context.Context, u *user.User) (*user.User, error) {
	_, err := r.db.ExecContext(ctx, `INSERT INTO users (id, username) VALUES ($1, $2)`, u.ID(), u.Username())
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("user create: %w", err)
	}
	return u, nil
}

type PGMessagesRepository struct {
	db *sql.DB
}

func NewPGMessagesRepository(db *sql.DB) *PGMessagesRepository {
	return &PGMessagesRepository{db: db}
}

func (r *PGMessagesRepository) Create(ctx context.Context, m *message.Message) (*message.Message, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("message create: %w", err)
	}
	defer tx.Rollback()

	var createdAt time.Time
	err = tx.QueryRowContext(ctx, `INSERT INTO messages (id, chat_id, sender_id, content) VALUES ($1, $2, $3, $4) RETURNING created_at`, m.ID(), m.ChatID(), m.SenderID(), m.Content()).Scan(&createdAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("message create: %w", err)
	}
	m.SetCreatedAt(createdAt)

	if _, err := tx.ExecContext(ctx, `UPDATE chats SET last_message_id = $1, last_message_sender_id = $2, last_message_content = $3, last_message_created_at = $4, message_count = message_count + 1 WHERE id = $5`, m.ID(), m.SenderID(), m.Content(), createdAt, m.ChatID()); err != nil {
		return nil, fmt.Errorf("message create: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("message create: %w", err)
	}
	return m, nil
}

func (r *PGMessagesRepository) ExistsInChat(ctx context.Context, messageID message.SnowflakeID, chatID chat.ChatID) (bool, error) {
	var messageExists int
	err := r.db.QueryRowContext(ctx, `SELECT 1 FROM messages WHERE id = $1 and chat_id = $2`, messageID, chatID).Scan(&messageExists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("messages existsInChat: %w", err)
	}

	return messageExists == 1, nil
}

func (r *PGMessagesRepository) UpdateLastReaded(ctx context.Context, readerID chat.ParticipantID, c *chat.Chat) error {
	lastReadedID, ok := c.LastReadMessageID(readerID)
	if !ok {
		return fmt.Errorf("messages updateLastReaded: no last read message")
	}

	var updated int
	err := r.db.QueryRowContext(ctx, `UPDATE chat_users SET last_read_message_id = $1 WHERE user_id = $2 and chat_id = $3 RETURNING 1`, lastReadedID, readerID, c.ID()).Scan(&updated)
	if err != nil {
		return fmt.Errorf("messages updateLastReaded: %w", err)
	}
	if updated != 1 {
		return fmt.Errorf("messages updateLastReaded: db returned something unexpected: %v", updated)
	}

	return nil
}
