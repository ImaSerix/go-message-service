package ws

import (
	"encoding/json"
	"slices"
	"sync"
	"time"

	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Hub хранит соединения пользователей. Карта connections читается и меняется
// только под mu. Соединение закрывается только после удаления из карты.
type Hub struct {
	connections map[uuid.UUID][]*Connection
	closed      bool
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
	conn := &Connection{
		id:   uuid,
		ws:   ws,
		Send: make(chan []byte, sendQueueSize),
	}

	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		ws.Close()
		return
	}
	h.connections[uuid] = append(h.connections[uuid], conn)
	h.mu.Unlock()

	go conn.readLoop(h.readLimit, h.pongWait, func() {
		h.removeConnection(conn)
	})

	go conn.writeLoop(h.writeWait, h.pingPeriod)
}

// removeConnection убирает из карты именно это соединение, а не все соединения
// пользователя, и закрывает его. Соединение, которого уже нет в карте, не трогается.
func (h *Hub) removeConnection(conn *Connection) {
	h.mu.Lock()
	conns := h.connections[conn.id]
	i := slices.Index(conns, conn)
	if i < 0 {
		h.mu.Unlock()
		return
	}
	conns = slices.Delete(conns, i, i+1)
	if len(conns) == 0 {
		delete(h.connections, conn.id)
	} else {
		h.connections[conn.id] = conns
	}
	h.mu.Unlock()

	conn.Close()
}

func (h *Hub) Close() {
	h.mu.Lock()
	connections := h.connections
	h.connections = make(map[uuid.UUID][]*Connection)
	h.closed = true
	h.mu.Unlock()

	for _, conns := range connections {
		for _, conn := range conns {
			conn.Close()
		}
	}
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

// sendRaw не ждёт медленного клиента: если очередь соединения полна,
// соединение закрывается.
func (h *Hub) sendRaw(id uuid.UUID, data []byte) {
	var overflowed []*Connection

	h.mu.RLock()
	for _, conn := range h.connections[id] {
		select {
		case conn.Send <- data:
		default:
			overflowed = append(overflowed, conn)
		}
	}
	h.mu.RUnlock()

	for _, conn := range overflowed {
		h.removeConnection(conn)
	}
}

func (h *Hub) SendToUsers(ids []uuid.UUID, event wsevent.WSEvent) {
	for _, id := range ids {
		h.SendToUser(id, event)
	}
}
