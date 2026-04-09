package message

import (
	"context"

	"github.com/google/uuid"
)

type ChatsRepository interface {
	UserIsParticipant(ctx context.Context, userID uuid.UUID, id uuid.UUID) (bool, error)
}
