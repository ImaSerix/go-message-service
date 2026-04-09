package chat

import (
	"context"
	"slices"
	"strings"

	"github.com/google/uuid"
)

type CreateChatInput struct {
	CallerID     uuid.UUID
	Name         string
	Participants []uuid.UUID
}

func (in CreateChatInput) Validate(ctx context.Context, users UsersRepository) (string, error) {
	if !slices.Contains(in.Participants, in.CallerID) {
		return "", ErrCallerIsNotParticipant
	}

	trimmedName := strings.TrimSpace(in.Name)
	if trimmedName == "" || len(trimmedName) < 2 || len(trimmedName) > 32 {
		return "", ErrInvalidName
	}
	return trimmedName, nil
}

type GetUserChatsInput struct {
	UserID uuid.UUID
}
