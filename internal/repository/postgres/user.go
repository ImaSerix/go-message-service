package postgres

import (
	"context"
	"database/sql"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type UsersRepository struct {
	db *sql.DB
}

func (ur *UsersRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	_, err := ur.db.ExecContext(ctx, `INSERT INTO users (id, username) VALUES ($1, $2)`, user.ID, user.Username)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return nil, repository.ErrAlreadyExists
			}
		}
		return nil, err
	}
	return user, nil
}
func (ur *UsersRepository) GetExistingIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	rows, err := ur.db.QueryContext(ctx, `SELECT id FROM users WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
