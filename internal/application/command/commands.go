package application

import "github.com/google/uuid"

type CreateChatCmd struct {
	CallerID     uuid.UUID
	Name         string
	Participants []uuid.UUID
}

type CreateMessageCmd struct {
	ChatID   uuid.UUID
	SenderID uuid.UUID
	Content  string
}

type ReadMessageCmd struct {
	ChatID    uuid.UUID
	CallerID  uuid.UUID
	MessageID int64
}

type CreateUserCmd struct {
	Username string
}
