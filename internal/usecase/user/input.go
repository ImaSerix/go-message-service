package user

import (
	"strings"

	"github.com/google/uuid"
)

type CreateInput struct {
	ID       uuid.UUID
	Username string
}

func (in *CreateInput) Validate() (string, error) {
	if in.ID == uuid.Nil {
		return "", ErrInvalidUUID
	}
	trimmedUsername := strings.TrimSpace(in.Username)
	if trimmedUsername == "" || len(trimmedUsername) < 2 || len(trimmedUsername) > 32 {
		return "", ErrInvalidUsername
	}
	return trimmedUsername, nil
}
