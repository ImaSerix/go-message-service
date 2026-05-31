package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ImaSerix/go-message-service/internal/application/command"
	"github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UsersRepository struct {
	db *sql.DB
}

func NewUsersRepository(db *sql.DB) *UsersRepository {
	return &UsersRepository{db: db}
}

func (r *UsersRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var userExists int
	err := r.db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE id = $1`, id).Scan(&userExists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("messages existsInChat: %w", err)
	}
	return userExists == 1, nil
}

func (r *UsersRepository) GetExistingIDs(ctx context.Context, ids []user.UserID) ([]user.UserID, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id FROM users WHERE id = ANY($1::uuid[])`, ids)
	if err != nil {
		return nil, fmt.Errorf("users GetExistingIDs: %w", err)
	}
	defer rows.Close()

	result := make([]user.UserID, 0)
	for rows.Next() {
		var id user.UserID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("users GetExistingIDs: %w", err)
		}
		result = append(result, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("users GetExistingIDs: %w", err)
	}

	return result, nil
}

func (r *UsersRepository) Create(ctx context.Context, u *user.User) (*user.User, error) {
	_, err := r.db.ExecContext(ctx, `INSERT INTO users (id, username) VALUES ($1, $2)`, u.ID(), u.Username())
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, command.ErrAlreadyExists
		}
		return nil, fmt.Errorf("user create: %w", err)
	}
	return u, nil
}
