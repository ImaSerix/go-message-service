package user

import "github.com/google/uuid"

type SaveLocalRequest struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
}
