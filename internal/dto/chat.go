package dto

import "github.com/google/uuid"

type Chat struct {
	ID           uuid.UUID
	Name         string
	Participants []uuid.UUID
	LastMessage  *Message
	UnreadCount  int
	MessageCount int
}
