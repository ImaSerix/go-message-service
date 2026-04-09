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

type UsersRepository struct {
	db *sql.DB
}

func NewUsersRepository(db *sql.DB) *UsersRepository {
	return &UsersRepository{
		db: db,
	}
}

func (ur *UsersRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {

	_, err := ur.db.ExecContext(ctx, `INSERT INTO users (id, username) VALUES ($1, $2)`, user.ID, user.Username)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return nil, repository.ErrAlreadyExists
			}
		}
		return nil, fmt.Errorf("user create: %w", err)
	}

	return user, nil
}

// GetExistingIDs метод репозитория, для определения какие пользователи в переданном списке существуют в базе данных
func (ur *UsersRepository) GetExistingIDs(ctx context.Context, userIDs []uuid.UUID) ([]uuid.UUID, error) {

	rows, err := ur.db.QueryContext(ctx, `SELECT id FROM users WHERE id = ANY($1)`, userIDs)
	if err != nil {
		return nil, fmt.Errorf("users GetExistingIDs: %w", err)
	}
	defer rows.Close()

	var result []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
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

func (ur *UsersRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {

	var userExists int
	err := ur.db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE id = $1`, id).Scan(&userExists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("user exists: %w", err)
	}
	if userExists != 1 {
		return false, fmt.Errorf("user exists: db returned something unexpected: %v", userExists)
	}

	return true, nil
}
