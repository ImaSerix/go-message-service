package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ChatsRepository struct {
	db *sql.DB
}

// Create stores chat and adds users as participants
func (cr *ChatsRepository) Create(ctx context.Context, chat *domain.Chat, participants []uuid.UUID) (*domain.Chat, error) {
	err := cr.db.QueryRowContext(ctx, `INSERT INTO chats (id, name) VALUES ($1, $2) RETURNING message_count, created_at`, chat.ID, chat.Name).Scan(&chat.MessageCount, &chat.CreatedAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return nil, repository.ErrAlreadyExists
			}
		}
		return nil, err
	}
	return chat, nil
}
func (cr *ChatsRepository) GetUserChats(ctx context.Context, userID uuid.UUID) ([]*domain.Chat, error) {
	query := `SELECT 
				id, 
				name, 
				message_count, 
				last_message_id, 
				last_message_sender_id, 
				last_message_content, 
				last_message_created_at, 
				created_at
			FROM chats 
			WHERE id IN (SELECT chat_id FROM users_chats WHERE user_id = $1);`
	rows, err := cr.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	var chats []*domain.Chat
	for rows.Next() {
		var chat domain.Chat
		if err := rows.Scan(&chat.ID, &chat.Name, &chat.MessageCount, &chat.LastMessage.ID, &chat.LastMessage.SenderID, &chat.LastMessage.Content, &chat.LastMessage.CreatedAt, &chat.CreatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}
func (cr *ChatsRepository) GetChatsParticipants(ctx context.Context, chatIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	rows, err := cr.db.QueryContext(ctx, `SELECT chat_id, user_id FROM users_chats WHERE chat_id = ANY($1)`, chatIDs)
	if err != nil {
		return nil, err
	}
	chatParticipants := make(map[uuid.UUID][]uuid.UUID)
	for rows.Next() {
		var chatID uuid.UUID
		var participantID uuid.UUID
		if err := rows.Scan(&chatID, &participantID); err != nil {
			return nil, err
		}
		chatParticipants[chatID] = append(chatParticipants[chatID], participantID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chatParticipants, nil
}
func (cr *ChatsRepository) UserIsParticipant(ctx context.Context, userID uuid.UUID, id uuid.UUID) (bool, error) {
	var isInChat int
	err := cr.db.QueryRowContext(ctx, `SELECT 1 FROM users_chats WHERE chat_id = $1 AND user_id = $2`, id, userID).Scan(&isInChat)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if isInChat != 1 {
		return false, fmt.Errorf("db returned something unexpected")
	}
	return true, nil
}
func (cr *ChatsRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var chatExists int
	err := cr.db.QueryRowContext(ctx, `SELECT 1 FROM chats WHERE chat_id = $1`, id).Scan(&chatExists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if chatExists != 1 {
		return false, fmt.Errorf("db returned something unexpected")
	}
	return true, nil
}
