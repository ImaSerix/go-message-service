package ws

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type client struct {
	id       uuid.UUID
	ws       *websocket.Conn
	lastSeen time.Time
}

type Hub struct {
	connections map[uuid.UUID][]*client
	mu          sync.RWMutex
}

func (h *Hub) AddConnection(uuid uuid.UUID, ws *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conn := &client{
		id:       uuid,
		ws:       ws,
		lastSeen: time.Now(),
	}

	h.connections[uuid] = append(h.connections[uuid], conn)
}

// TODO Нужно сделать хаб
