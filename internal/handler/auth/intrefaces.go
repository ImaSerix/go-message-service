package auth

import (
	"context"

	"github.com/google/uuid"
)

type UsersRepository interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}
