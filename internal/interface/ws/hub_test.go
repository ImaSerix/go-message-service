package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ImaSerix/go-message-service/internal/interface/handler/auth"
	wsHandler "github.com/ImaSerix/go-message-service/internal/interface/handler/ws"
	"github.com/ImaSerix/go-message-service/internal/wsevent"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func newTestHub() *Hub {
	return New(512, time.Minute, 50*time.Second, 10*time.Second)
}

// newTestServer отдаёт маршрут GET /ws, как в cmd/main.go; пользователь берётся из ?user=.
func newTestServer(t *testing.T, hub *Hub) *httptest.Server {
	t.Helper()
	h := wsHandler.New(hub)
	r := chi.NewRouter()
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		id := uuid.MustParse(r.URL.Query().Get("user"))
		h.ServeHTTP(w, r.WithContext(auth.WithUserID(r.Context(), id)))
	})
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

func dial(t *testing.T, srv *httptest.Server, user uuid.UUID) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?user=" + user.String()
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func connCount(h *Hub, user uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections[user])
}

func waitConnCount(t *testing.T, h *Hub, user uuid.UUID, n int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for connCount(h, user) != n {
		if time.Now().After(deadline) {
			t.Fatalf("expected %d connections of user, got %d", n, connCount(h, user))
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func newMessageEvent(content string) *wsevent.NewMessage {
	return &wsevent.NewMessage{Message: wsevent.Message{ID: 1, Content: content}}
}

// readNewMessage читает события, пока не придёт new_message с текстом content.
func readNewMessage(t *testing.T, conn *websocket.Conn, content string) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		var env struct {
			Type string             `json:"type"`
			Data wsevent.NewMessage `json:"data"`
		}
		if err := json.Unmarshal(data, &env); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if env.Type != "new_message" {
			t.Fatalf("expected type new_message, got %q", env.Type)
		}
		if env.Data.Message.Content == content {
			return
		}
	}
}

func expectClosed(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
				t.Fatal("expected connection to be closed, but it is still open")
			}
			return
		}
	}
}

func TestHub_UpgradeAndNewMessage(t *testing.T) {
	hub := newTestHub()
	srv := newTestServer(t, hub)
	user := uuid.New()

	conn := dial(t, srv, user)
	waitConnCount(t, hub, user, 1)

	hub.SendToUser(user, newMessageEvent("hello"))
	readNewMessage(t, conn, "hello")
}

func TestHub_CloseOneOfTwoTabsDuringSend(t *testing.T) {
	hub := newTestHub()
	srv := newTestServer(t, hub)
	user := uuid.New()

	first := dial(t, srv, user)
	second := dial(t, srv, user)
	waitConnCount(t, hub, user, 2)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 5 {
			hub.SendToUser(user, newMessageEvent("during close"))
		}
	}()
	first.Close()
	<-done

	waitConnCount(t, hub, user, 1)
	hub.SendToUser(user, newMessageEvent("after close"))
	readNewMessage(t, second, "after close")
}

func TestHub_SlowClientOverflowDoesNotBlockOthers(t *testing.T) {
	hub := newTestHub()
	srv := newTestServer(t, hub)
	slowUser, otherUser := uuid.New(), uuid.New()

	// Соединение медленного клиента без writeLoop: из его очереди никто не читает.
	serverSide := make(chan *websocket.Conn, 1)
	upgradeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		serverSide <- c
	}))
	t.Cleanup(upgradeSrv.Close)
	slowClient, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(upgradeSrv.URL, "http"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { slowClient.Close() })
	slow := &Connection{id: slowUser, ws: <-serverSide, Send: make(chan []byte, sendQueueSize)}
	hub.mu.Lock()
	hub.connections[slowUser] = []*Connection{slow}
	hub.mu.Unlock()

	other := dial(t, srv, otherUser)
	waitConnCount(t, hub, otherUser, 1)

	for range sendQueueSize {
		hub.SendToUser(slowUser, newMessageEvent("fill"))
	}
	if connCount(hub, slowUser) != 1 {
		t.Fatal("full queue must not close the connection yet")
	}

	sent := make(chan struct{})
	go func() {
		hub.SendToUsers([]uuid.UUID{slowUser, otherUser}, newMessageEvent("overflow"))
		close(sent)
	}()
	select {
	case <-sent:
	case <-time.After(2 * time.Second):
		t.Fatal("send blocked on slow client")
	}

	if connCount(hub, slowUser) != 0 {
		t.Fatal("expected slow connection to be removed after overflow")
	}
	for range sendQueueSize {
		<-slow.Send
	}
	if _, ok := <-slow.Send; ok {
		t.Fatal("expected send queue of slow connection to be closed")
	}
	expectClosed(t, slowClient)
	readNewMessage(t, other, "overflow")
}

func TestHub_Shutdown(t *testing.T) {
	hub := newTestHub()
	srv := newTestServer(t, hub)
	user := uuid.New()

	first := dial(t, srv, user)
	second := dial(t, srv, user)
	waitConnCount(t, hub, user, 2)

	hub.Close()
	hub.Close()
	hub.SendToUser(user, newMessageEvent("after shutdown"))

	expectClosed(t, first)
	expectClosed(t, second)

	late := dial(t, srv, user)
	expectClosed(t, late)
	if connCount(hub, user) != 0 {
		t.Fatal("expected no connections after shutdown")
	}
}
