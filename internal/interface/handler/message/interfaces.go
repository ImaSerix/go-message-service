package message

import (
	"context"

	"github.com/google/uuid"
)

type ChatsQry interface {
	UserIsParticipant(ctx context.Context, userID uuid.UUID, id uuid.UUID) (bool, error)
}
