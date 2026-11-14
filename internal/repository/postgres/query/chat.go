package query

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ImaSerix/go-message-service/internal/application/query"
	"github.com/google/uuid"
)

type ChatsQry struct {
	db *sql.DB
}

func NewChatsQry(db *sql.DB) *ChatsQry {
	return &ChatsQry{
		db: db,
	}
}

func (q *ChatsQry) GetChat(ctx context.Context, chatID uuid.UUID) (*query.GetUserOutput, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("chat get: %w", err)
	}
	defer tx.Rollback()

	qry := `SELECT
					id,
					name,
					message_count,
					created_at
				FROM chats
				WHERE id = $1;`
	c := &query.GetUserOutput{}
	err = tx.QueryRowContext(ctx, qry, chatID).Scan(&c.ID, &c.Name, &c.MessageCount, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("chat get: %w", err)
	}

	qry = `SELECT
					user_id
				FROM chat_users
				WHERE chat_id = $1;`
	rows, err := tx.QueryContext(ctx, qry, chatID)
	if err != nil {
		return nil, fmt.Errorf("chat get: %w", err)
	}
	defer rows.Close()

	participants := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(
			&id,
		); err != nil {
			return nil, fmt.Errorf("chat get: %w", err)
		}
		participants = append(participants, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chat get: %w", err)
	}

	c.Participants = participants
	return c, nil
}

func (q *ChatsQry) GetUserChats(ctx context.Context, userID uuid.UUID) ([]*query.GetUserChatsOutput, error) {

	qry := `SELECT 
					c.id, 
					c.name, 
					c.last_message_id, 
					c.last_message_sender_id, 
					c.last_message_content, 
					c.last_message_created_at,
					cu.last_read_message_id,
					c.created_at
				FROM chat_users cu
				INNER JOIN chats c
				ON cu.chat_id = c.id
				WHERE cu.user_id = $1;`
	rows, err := q.db.QueryContext(ctx, qry, userID)
	if err != nil {
		return nil, fmt.Errorf("chats get: %w", err)
	}
	defer rows.Close()

	res := []*query.GetUserChatsOutput{}
	for rows.Next() {
		c := query.GetUserChatsOutput{}
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.LastMessageID,
			&c.LastMessageSenderID,
			&c.LastMessageContent,
			&c.LastMessageCreatedAt,
			&c.LastReadMessageID,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("chats get: %w", err)
		}
		res = append(res, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chats get: %w", err)
	}
	return res, nil
}
