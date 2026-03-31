package wsevent

import (
	"github.com/ImaSerix/go-message-service/internal/ws"
)

type WSEvent interface {
	Name() string
}

type NewChat struct {
	Chat ws.Chat `json:"chat"`
}

func (e *NewChat) Name() string {
	return "new_chat"
}

type NewMessage struct {
	Message ws.Message `json:"message"`
}

func (e *NewMessage) Name() string {
	return "new_message"
}
