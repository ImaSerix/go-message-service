package application

import "github.com/google/uuid"

type CreateChatCmd struct {
	CallerID     uuid.UUID
	Name         string
	Participants []uuid.UUID
}
