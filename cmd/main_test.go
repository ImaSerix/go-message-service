package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ImaSerix/go-message-service/internal/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// parseTestDatabase validates the effective database before any connection or cleanup.
func parseTestDatabase(raw string) (*pgx.ConnConfig, error) {
	cfg, err := pgx.ParseConfig(raw)
	if err != nil {
		return nil, errors.New("TEST_DATABASE_URL has invalid PostgreSQL settings")
	}
	if !strings.HasSuffix(cfg.Database, "_test") {
		return nil, errors.New("TEST_DATABASE_URL database name must end with _test")
	}
	return cfg, nil
}

// newTestDatabase returns a connection to an emptied test database; migrations must already be applied.
// Tests using this helper must not run in parallel because they share tables.
func newTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	cfg, err := parseTestDatabase(raw)
	if err != nil {
		t.Fatal(err)
	}
	db := stdlib.OpenDB(*cfg)
	t.Cleanup(func() { db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatal("test database is unavailable")
	}
	if _, err := db.Exec("TRUNCATE chat_users, messages, chats, users"); err != nil {
		t.Fatal("could not reset test tables; apply migrations first")
	}
	return db
}

func TestDatabaseGuard(t *testing.T) {
	tests := []struct {
		raw   string
		valid bool
	}{
		{"postgres://user:pass@127.0.0.1/message_test?sslmode=disable", true},
		{"host=127.0.0.1 dbname=message_test", true},
		{"postgres://user@127.0.0.1/message", false},
		{"postgres://user@127.0.0.1/message_test_copy", false},
		{"postgres://user@127.0.0.1/message_test?dbname=production", false},
		{"postgres://user@127.0.0.1/production?dbname=message_test", true},
		{"://invalid", false},
	}
	for _, tc := range tests {
		_, err := parseTestDatabase(tc.raw)
		if (err == nil) != tc.valid {
			t.Errorf("guard for %q: valid = %v; want %v", tc.raw, err == nil, tc.valid)
		}
	}
}

// do sends a request to the router and decodes a JSON response body into out, if out is not nil.
func do(t *testing.T, router http.Handler, method, path, userID, body string, wantStatus int, out any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if userID != "" {
		req.Header.Set("X-User-ID", userID)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("%s %s: status = %d, body = %q; want %d", method, path, rec.Code, rec.Body.String(), wantStatus)
	}
	if out == nil {
		return
	}
	if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
		t.Fatalf("%s %s: decode body %q: %v", method, path, rec.Body.String(), err)
	}
}

type chatsResponse struct {
	Chats []struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		LastMessage *struct {
			ID       int64     `json:"id"`
			SenderID uuid.UUID `json:"sender_id"`
			Content  string    `json:"content"`
		} `json:"last_message"`
		UnreadCount int `json:"unread_count"`
	} `json:"chats"`
}

// TestChatScenario runs the main HTTP scenario from README against a real PostgreSQL.
func TestChatScenario(t *testing.T) {
	db := newTestDatabase(t)
	router := newRouter(&config.Config{
		WSHub: config.WSHubConfig{
			ReadLimit:  512,
			PongWait:   60 * time.Second,
			PingPerion: 54 * time.Second,
			WriteWait:  10 * time.Second,
		},
		IDGenerator:        config.IDGeneratorConfig{NodeID: 0, Epoch: 1700000000},
		MaxMessagePageSize: 100,
	}, db)

	alice, bob, stranger := uuid.NewString(), uuid.NewString(), uuid.NewString()

	// Users.
	do(t, router, http.MethodPost, "/internal/users", "", fmt.Sprintf(`{"user_id":%q,"username":"alice"}`, alice), http.StatusCreated, nil)
	do(t, router, http.MethodPost, "/internal/users", "", fmt.Sprintf(`{"user_id":%q,"username":"bob"}`, bob), http.StatusCreated, nil)
	do(t, router, http.MethodPost, "/internal/users", "", fmt.Sprintf(`{"user_id":%q,"username":"alice"}`, alice), http.StatusConflict, nil)

	// A user that does not exist in Message is not authorized.
	do(t, router, http.MethodGet, "/chats", stranger, "", http.StatusUnauthorized, nil)
	do(t, router, http.MethodPost, "/chats", stranger, fmt.Sprintf(`{"name":"chat","participants":[%q]}`, stranger), http.StatusUnauthorized, nil)

	// No chats yet.
	var chats chatsResponse
	do(t, router, http.MethodGet, "/chats", alice, "", http.StatusOK, &chats)
	if len(chats.Chats) != 0 {
		t.Fatalf("chats before create = %d; want 0", len(chats.Chats))
	}

	// A participant that does not exist in Message is rejected.
	do(t, router, http.MethodPost, "/chats", alice, fmt.Sprintf(`{"name":"chat","participants":[%q,%q]}`, alice, stranger), http.StatusBadRequest, nil)

	// The caller must be one of the participants.
	do(t, router, http.MethodPost, "/chats", alice, fmt.Sprintf(`{"name":"chat","participants":[%q]}`, bob), http.StatusForbidden, nil)

	// Create chat.
	var created struct {
		ID uuid.UUID `json:"id"`
	}
	do(t, router, http.MethodPost, "/chats", alice, fmt.Sprintf(`{"name":"chat","participants":[%q,%q]}`, alice, bob), http.StatusCreated, &created)
	if created.ID == uuid.Nil {
		t.Fatal("created chat id is empty")
	}
	chatPath := "/chats/" + created.ID.String()

	// Chat without messages.
	do(t, router, http.MethodGet, "/chats", bob, "", http.StatusOK, &chats)
	if len(chats.Chats) != 1 || chats.Chats[0].ID != created.ID || chats.Chats[0].LastMessage != nil || chats.Chats[0].UnreadCount != 0 {
		t.Fatalf("chats before message = %+v; want one chat %v without messages", chats.Chats, created.ID)
	}
	var history struct {
		Messages []struct {
			ID       int64     `json:"id"`
			SenderID uuid.UUID `json:"sender_id"`
			Content  string    `json:"content"`
		} `json:"messages"`
	}
	do(t, router, http.MethodGet, chatPath+"/messages", bob, "", http.StatusOK, &history)
	if len(history.Messages) != 0 {
		t.Fatalf("history before message = %d; want 0", len(history.Messages))
	}

	// Only participants have access to the chat.
	do(t, router, http.MethodGet, chatPath, stranger, "", http.StatusUnauthorized, nil)
	do(t, router, http.MethodPost, "/internal/users", "", fmt.Sprintf(`{"user_id":%q,"username":"carol"}`, stranger), http.StatusCreated, nil)
	do(t, router, http.MethodGet, chatPath, stranger, "", http.StatusForbidden, nil)
	do(t, router, http.MethodGet, "/chats/"+uuid.NewString(), alice, "", http.StatusForbidden, nil)

	// Send messages.
	var first, second struct {
		ID int64 `json:"id"`
	}
	do(t, router, http.MethodPost, chatPath+"/messages", alice, `{"content":"hello"}`, http.StatusCreated, &first)
	do(t, router, http.MethodPost, chatPath+"/messages", alice, `{"content":"world"}`, http.StatusCreated, &second)
	if first.ID == 0 || second.ID <= first.ID {
		t.Fatalf("message ids = %d, %d; want increasing non-zero ids", first.ID, second.ID)
	}

	// Chat list.
	do(t, router, http.MethodGet, "/chats", bob, "", http.StatusOK, &chats)
	if len(chats.Chats) != 1 {
		t.Fatalf("chats = %d; want 1", len(chats.Chats))
	}
	c := chats.Chats[0]
	if c.ID != created.ID || c.Name != "chat" || c.UnreadCount != 2 {
		t.Fatalf("chat in list = %+v; want id %v, name chat, unread 2", c, created.ID)
	}
	if c.LastMessage == nil || c.LastMessage.ID != second.ID || c.LastMessage.SenderID.String() != alice || c.LastMessage.Content != "world" {
		t.Fatalf("last message = %+v; want id %d from %s", c.LastMessage, second.ID, alice)
	}

	// Chat details.
	var details struct {
		ID           uuid.UUID   `json:"id"`
		Name         string      `json:"name"`
		MessageCount int         `json:"message_count"`
		Participants []uuid.UUID `json:"participants"`
	}
	do(t, router, http.MethodGet, chatPath, bob, "", http.StatusOK, &details)
	if details.ID != created.ID || details.Name != "chat" || details.MessageCount != 2 || len(details.Participants) != 2 {
		t.Fatalf("chat details = %+v; want id %v, name chat, 2 messages, 2 participants", details, created.ID)
	}

	// History. README does not define the order of a page without a cursor, so it is not checked here.
	do(t, router, http.MethodGet, chatPath+"/messages", bob, "", http.StatusOK, &history)
	if len(history.Messages) != 2 || history.Messages[0].ID+history.Messages[1].ID != first.ID+second.ID || history.Messages[0].ID == history.Messages[1].ID {
		t.Fatalf("history = %+v; want messages %d and %d", history.Messages, first.ID, second.ID)
	}
	do(t, router, http.MethodGet, fmt.Sprintf("%s/messages?after=%d", chatPath, first.ID), bob, "", http.StatusOK, &history)
	if len(history.Messages) != 1 || history.Messages[0].ID != second.ID {
		t.Fatalf("history after %d = %+v; want message %d", first.ID, history.Messages, second.ID)
	}
	do(t, router, http.MethodGet, fmt.Sprintf("%s/messages?before=%d", chatPath, second.ID), bob, "", http.StatusOK, &history)
	if len(history.Messages) != 1 || history.Messages[0].ID != first.ID {
		t.Fatalf("history before %d = %+v; want message %d", second.ID, history.Messages, first.ID)
	}
	if history.Messages[0].SenderID.String() != alice || history.Messages[0].Content != "hello" {
		t.Fatalf("first message = %+v; want hello from %s", history.Messages[0], alice)
	}
	do(t, router, http.MethodGet, chatPath+"/messages?after=1", bob, "", http.StatusBadRequest, nil)

	// Mark as read.
	do(t, router, http.MethodPost, chatPath+"/read", bob, fmt.Sprintf(`{"last_read_message_id":%d}`, first.ID), http.StatusNoContent, nil)
	do(t, router, http.MethodGet, "/chats", bob, "", http.StatusOK, &chats)
	if len(chats.Chats) != 1 || chats.Chats[0].UnreadCount != 1 {
		t.Fatalf("chats after first read = %+v; want unread 1", chats.Chats)
	}
	do(t, router, http.MethodPost, chatPath+"/read", bob, fmt.Sprintf(`{"last_read_message_id":%d}`, second.ID), http.StatusNoContent, nil)
	do(t, router, http.MethodGet, "/chats", bob, "", http.StatusOK, &chats)
	if len(chats.Chats) != 1 || chats.Chats[0].UnreadCount != 0 {
		t.Fatalf("chats after second read = %+v; want unread 0", chats.Chats)
	}
}
