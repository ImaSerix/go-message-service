package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ImaSerix/go-message-service/internal/application/command"
	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ChatsRepository struct {
	db *sql.DB
}

func NewChatsRepository(db *sql.DB) *ChatsRepository {
	return &ChatsRepository{db: db}
}

func (r *ChatsRepository) Create(ctx context.Context, c *chat.Chat) (*chat.Chat, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `INSERT INTO chats (id, name) VALUES ($1, $2)`, c.ID(), c.Name())
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, command.ErrAlreadyExists
		}
		return nil, fmt.Errorf("create chat: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO chat_users (chat_id, user_id) SELECT $1, unnest($2::uuid[])`, c.ID(), c.Participants()); err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	return c, nil
}

func (r *ChatsRepository) LoadForSend(ctx context.Context, id chat.ChatID) (*chat.Chat, error) {
	return r.loadChat(ctx, id)
}

func (r *ChatsRepository) LoadForRead(ctx context.Context, id chat.ChatID) (*chat.Chat, error) {
	return r.loadChat(ctx, id)
}

func (r *ChatsRepository) loadChat(ctx context.Context, id chat.ChatID) (*chat.Chat, error) {
	var name string
	err := r.db.QueryRowContext(ctx, `SELECT name FROM chats WHERE id = $1`, id).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, command.ErrNotFound
		}
		return nil, fmt.Errorf("load chat: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `SELECT user_id, last_read_message_id FROM chat_users WHERE chat_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("load chat: %w", err)
	}
	defer rows.Close()

	participants := make([]chat.ParticipantID, 0)
	lastReadMessages := make(map[chat.ParticipantID]*chat.SnowflakeID)
	for rows.Next() {
		var participant chat.ParticipantID
		var lastReadMessageID *chat.SnowflakeID
		if err := rows.Scan(&participant, &lastReadMessageID); err != nil {
			return nil, fmt.Errorf("load chat: %w", err)
		}
		participants = append(participants, participant)
		lastReadMessages[participant] = lastReadMessageID
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

	return chat.RehydrateChat(id, chatName, participants, nil, lastReadMessages), nil
}

func (r *ChatsRepository) UserIsParticipant(ctx context.Context, userID uuid.UUID, chatID uuid.UUID) (bool, error) {
	var userIsParticipant int
	err := r.db.QueryRowContext(ctx, `SELECT 1 FROM chat_users WHERE chat_id = $1 and user_id = $2`, chatID, userID).Scan(&userIsParticipant)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("chats userIsParticipant: %w", err)
	}

	return userIsParticipant == 1, nil
}
