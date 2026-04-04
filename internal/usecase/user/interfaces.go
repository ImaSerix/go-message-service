package user

import (
	"context"

	"github.com/ImaSerix/go-message-service/internal/domain"
)

type UsersRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
}
