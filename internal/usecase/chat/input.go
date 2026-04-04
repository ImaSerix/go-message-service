package chat

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type CreateChatInput struct {
	Name         string
	Participants []uuid.UUID
}

func (in CreateChatInput) Validate(ctx context.Context, users UsersRepository) (string, error) {
	trimmedName := strings.TrimSpace(in.Name)
	if trimmedName == "" || len(trimmedName) < 2 || len(trimmedName) > 32 {
		return "", ErrInvalidName
	}
	return trimmedName, nil
}

type GetUserChatsInput struct {
	UserID uuid.UUID
}
