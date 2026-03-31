package user

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/dto"
	"github.com/ImaSerix/go-message-service/internal/repository"
	"github.com/google/uuid"
)

type UsersUseCase struct {
	users UsersRepository
}

type CreateInput struct {
	ID       uuid.UUID
	Username string
}

func (uuc *UsersUseCase) Create(in *CreateInput) (*dto.User, error) {
	if in.ID == uuid.Nil {
		return nil, ErrInvalidUUID
	}
	trimmedUsername := strings.TrimSpace(in.Username)
	if trimmedUsername == "" || len(trimmedUsername) < 2 || len(trimmedUsername) > 32 {
		return nil, ErrInvalidUsername
	}

	user := &domain.User{
		ID:       uuid.New(),
		Username: trimmedUsername,
	}
	user, err := uuc.users.Create(user)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("create: %w", err)
	}

	return &dto.User{
		ID:       user.ID,
		Username: user.Username,
	}, nil
}
