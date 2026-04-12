package ws

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WSHub interface {
	AddConnection(uuid uuid.UUID, ws *websocket.Conn)
}
