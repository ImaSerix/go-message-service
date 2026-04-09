package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ChatsRepository struct {
	db *sql.DB
}

func NewChatsRepository(db *sql.DB) *ChatsRepository {
	return &ChatsRepository{
		db: db,
	}
}

// Create stores chat and inserts users as participants
func (cr *ChatsRepository) Create(ctx context.Context, chat *domain.Chat, participants []uuid.UUID) (*domain.Chat, error) {

	if len(participants) == 0 {
		return nil, repository.ErrEmptyParticipants
	}

	tx, err := cr.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `INSERT INTO chats (id, name) VALUES ($1, $2) RETURNING *`, chat.ID, chat.Name).
		Scan(
			&chat.ID,
			&chat.Name,
			&chat.MessageCount,
			&chat.LastMessage.ID,
			&chat.LastMessage.SenderID,
			&chat.LastMessage.Content,
			&chat.LastMessage.CreatedAt,
			&chat.CreatedAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return nil, repository.ErrAlreadyExists
			}
		}
		return nil, fmt.Errorf("create chat: %w", err)
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO chats_users (chat_id, user_id) SELECT $1, unnest($2::uuid[]) ON CONFLICT DO NOTHING`, chat.ID, participants)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
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
		return nil, fmt.Errorf("getUserChats: %w", err)
	}
	defer rows.Close()

	chats := make([]*domain.Chat, 0)
	for rows.Next() {
		var chat domain.Chat
		var lastMessageID *domain.SnowflakeID
		var senderID *uuid.UUID
		var content *string
		var createdAt *time.Time
		if err := rows.Scan(
			&chat.ID,
			&chat.Name,
			&chat.MessageCount,
			&lastMessageID,
			&senderID,
			&content,
			&createdAt,
			&chat.CreatedAt); err != nil {
			return nil, fmt.Errorf("getUserChats: %w", err)
		}
		if lastMessageID != nil {
			chat.LastMessage = &domain.Message{
				ID:        *lastMessageID,
				ChatID:    chat.ID,
				SenderID:  *senderID,
				Content:   *content,
				CreatedAt: *createdAt,
			}
		}
		chats = append(chats, &chat)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("getUserChats: %w", err)
	}

	return chats, nil
}

// TODO: Этого нет в obsidian, также там парочку вещей которые используют middleware
func (cr *ChatsRepository) GetChatParticipants(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := cr.db.QueryContext(ctx, `SELECT user_id FROM users_chats WHERE chat_id = $1`, chatID)
	if err != nil {
		return nil, fmt.Errorf("getChatParticipants: %w", err)
	}
	defer rows.Close()

	chatParticipants := []uuid.UUID{}
	for rows.Next() {
		var participantID uuid.UUID
		if err := rows.Scan(&participantID); err != nil {
			return nil, fmt.Errorf("getChatParticipants: %w", err)
		}
		chatParticipants = append(chatParticipants, participantID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("getChatParticipants: %w", err)
	}

	return chatParticipants, nil
}

func (cr *ChatsRepository) GetChatsParticipants(ctx context.Context, chatIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {

	rows, err := cr.db.QueryContext(ctx, `SELECT chat_id, user_id FROM users_chats WHERE chat_id = ANY($1)`, chatIDs)
	if err != nil {
		return nil, fmt.Errorf("getChatsParticipants: %w", err)
	}
	defer rows.Close()

	chatParticipants := make(map[uuid.UUID][]uuid.UUID)
	for rows.Next() {
		var chatID uuid.UUID
		var participantID uuid.UUID
		if err := rows.Scan(&chatID, &participantID); err != nil {
			return nil, fmt.Errorf("getChatsParticipants: %w", err)
		}
		chatParticipants[chatID] = append(chatParticipants[chatID], participantID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("getChatsParticipants: %w", err)
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
		return false, fmt.Errorf("chat userIsParticioants: %w", err)
	}
	if isInChat != 1 {
		return false, fmt.Errorf("chat userIsParticioants: db returned something unexpected: %v", isInChat)
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
		return false, fmt.Errorf("chat exists: %w", err)
	}
	if chatExists != 1 {
		return false, fmt.Errorf("chat exists: db returned something unexpected: %v", chatExists)
	}

	return true, nil
}
