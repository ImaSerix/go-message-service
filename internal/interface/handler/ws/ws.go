package ws

import (
	"log/slog"
	"net/http"

	"github.com/ImaSerix/go-message-service/internal/interface/handler/auth"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	hub WSHub
}

func New(hub WSHub) *WSHandler {
	return &WSHandler{
		hub: hub,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// TODO: Может быть надо ограничить, или упоменуть в Readme об этом
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		slog.Error("reading user from context failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.hub.AddConnection(userID, conn)
}
