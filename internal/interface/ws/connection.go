package ws

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// sendQueueSize — размер очереди отправки одного соединения. При переполнении
// соединение закрывается, клиент переподключается и перечитывает историю по HTTP.
const sendQueueSize = 16

type Connection struct {
	id        uuid.UUID
	ws        *websocket.Conn
	Send      chan []byte
	closeOnce sync.Once
}

func (conn *Connection) readLoop(readLimit int64, pongWait time.Duration, onClose func()) {
	defer onClose()
	conn.ws.SetReadLimit(readLimit)
	conn.ws.SetReadDeadline(time.Now().Add(pongWait))

	conn.ws.SetPongHandler(func(appData string) error {
		conn.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		if _, _, err := conn.ws.ReadMessage(); err != nil {
			return
		}
	}
}

func (conn *Connection) writeLoop(writeWait time.Duration, pingPeriod time.Duration) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	// Ошибка записи закрывает сокет: readLoop получает ошибку и убирает соединение из Hub.
	defer conn.ws.Close()

	for {
		select {
		case msg, ok := <-conn.Send:
			if !ok {
				return
			}

			conn.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			conn.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Close вызывает только Hub, после того как соединение убрано из карты:
// тогда в Send больше никто не пишет. Повторный вызов ничего не делает.
func (conn *Connection) Close() {
	conn.closeOnce.Do(func() {
		conn.ws.Close()
		close(conn.Send)
	})
}
