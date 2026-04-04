package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/dto"
	"github.com/ImaSerix/go-message-service/internal/repository"
	"github.com/google/uuid"
)

type UsersUseCase struct {
	users UsersRepository
}

func (uuc *UsersUseCase) Create(ctx context.Context, in *CreateInput) (*dto.User, error) {

	trimmedUsername, err := in.Validate()
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:       uuid.New(),
		Username: trimmedUsername,
	}

	user, err = uuc.users.Create(ctx, user)
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
