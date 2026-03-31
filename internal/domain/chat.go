package domain

import (
	"github.com/google/uuid"
)

type Chat struct {
	ID           uuid.UUID
	Name         string
	LastMessage  *Message
	MessageCount int
}
