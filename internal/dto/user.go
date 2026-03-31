package dto

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID
	Username string
}
