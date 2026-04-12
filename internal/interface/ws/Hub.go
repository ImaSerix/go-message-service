package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Hub struct {
	connections map[uuid.UUID][]*Connection
	mu          sync.RWMutex
	readLimit   int64
	pongWait    time.Duration
	pingPeriod  time.Duration
	writeWait   time.Duration
}

func New(readLimit int64, pongWait time.Duration, pingPeriod time.Duration, writeWait time.Duration) *Hub {
	return &Hub{
		connections: map[uuid.UUID][]*Connection{},
		mu:          sync.RWMutex{},
		readLimit:   readLimit,
		pongWait:    pongWait,
		pingPeriod:  pingPeriod,
		writeWait:   writeWait,
	}
}

func (h *Hub) AddConnection(uuid uuid.UUID, ws *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conn := &Connection{
		id:   uuid,
		ws:   ws,
		Send: make(chan []byte),
	}

	go conn.readLoop(h.readLimit, h.pongWait, func() {
		conn.Close()
		delete(h.connections, conn.id)
	})

	go conn.writeLoop(h.writeWait, h.pingPeriod)

	h.connections[uuid] = append(h.connections[uuid], conn)
}

func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, conns := range h.connections {
		for _, conn := range conns {
			conn.Close()
		}
	}

	h.connections = make(map[uuid.UUID][]*Connection)
}

func (h *Hub) SendToUser(id uuid.UUID, event wsevent.WSEvent) {
	payload := map[string]any{
		"type": event.Name(),
		"data": event,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	h.sendRaw(id, data)
}

func (h *Hub) sendRaw(id uuid.UUID, data []byte) {
	conns, ok := h.connections[id]
	if !ok {
		return
	}

	for _, conn := range conns {
		conn.Send <- data
	}
}

func (h *Hub) SendToUsers(ids []uuid.UUID, event wsevent.WSEvent) {
	for _, id := range ids {
		h.SendToUser(id, event)
	}
}
