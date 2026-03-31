package user

import "github.com/ImaSerix/go-message-service/internal/domain"

type UsersRepository interface {
	Create(user *domain.User) (*domain.User, error)
}
