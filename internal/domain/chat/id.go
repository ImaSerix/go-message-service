package chat

import "github.com/google/uuid"

type ChatID uuid.UUID

type ParticipantID uuid.UUID

type SnowflakeID int64

// Older возвращает true, если вызываемый ID старше, чем передаваемый
func (id SnowflakeID) Older(otherID SnowflakeID) bool {
	return id < otherID
}
