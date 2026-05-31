package query

import (
	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/google/uuid"
)

type GetUserChatsQry struct {
	UserID uuid.UUID
}

type GetUserMessagesQry struct {
	ChatID    uuid.UUID
	UserID    uuid.UUID
	Cursor    *int64
	Direction domain.Direction
	Limit     int
}

type GetUserChatQry struct {
	UserID uuid.UUID
	ChatID uuid.UUID
}
