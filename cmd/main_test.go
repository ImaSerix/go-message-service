package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ImaSerix/go-message-service/internal/config"
	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/message"
	"github.com/ImaSerix/go-message-service/internal/repository/postgres/command"
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

	// Repeated participants are removed silently.
	var details struct {
		ID           uuid.UUID   `json:"id"`
		Name         string      `json:"name"`
		MessageCount int         `json:"message_count"`
		Participants []uuid.UUID `json:"participants"`
	}
	var other struct {
		ID uuid.UUID `json:"id"`
	}
	do(t, router, http.MethodPost, "/chats", alice, fmt.Sprintf(`{"name":"other","participants":[%q,%q,%q,%q]}`, alice, stranger, alice, stranger), http.StatusCreated, &other)
	otherPath := "/chats/" + other.ID.String()
	do(t, router, http.MethodGet, otherPath, alice, "", http.StatusOK, &details)
	if got, want := sortedIDs(details.Participants), sortedIDs([]uuid.UUID{uuid.MustParse(alice), uuid.MustParse(stranger)}); !slices.Equal(got, want) {
		t.Fatalf("participants of chat with repeated participants = %v; want %v", got, want)
	}
	var foreign struct {
		ID int64 `json:"id"`
	}
	do(t, router, http.MethodPost, otherPath+"/messages", alice, `{"content":"elsewhere"}`, http.StatusCreated, &foreign)

	// Send messages.
	var first, second, third struct {
		ID int64 `json:"id"`
	}
	do(t, router, http.MethodPost, chatPath+"/messages", alice, `{"content":"hello"}`, http.StatusCreated, &first)
	do(t, router, http.MethodPost, chatPath+"/messages", alice, `{"content":"world"}`, http.StatusCreated, &second)
	do(t, router, http.MethodPost, chatPath+"/messages", alice, `{"content":"again"}`, http.StatusCreated, &third)
	if first.ID == 0 || second.ID <= first.ID || third.ID <= second.ID {
		t.Fatalf("message ids = %d, %d, %d; want increasing non-zero ids", first.ID, second.ID, third.ID)
	}

	// Chat list.
	do(t, router, http.MethodGet, "/chats", bob, "", http.StatusOK, &chats)
	if len(chats.Chats) != 1 {
		t.Fatalf("chats = %d; want 1", len(chats.Chats))
	}
	c := chats.Chats[0]
	if c.ID != created.ID || c.Name != "chat" || c.UnreadCount != 3 {
		t.Fatalf("chat in list = %+v; want id %v, name chat, unread 3", c, created.ID)
	}
	if c.LastMessage == nil || c.LastMessage.ID != third.ID || c.LastMessage.SenderID.String() != alice || c.LastMessage.Content != "again" {
		t.Fatalf("last message = %+v; want id %d from %s", c.LastMessage, third.ID, alice)
	}

	// Chat details.
	do(t, router, http.MethodGet, chatPath, bob, "", http.StatusOK, &details)
	if details.ID != created.ID || details.Name != "chat" || details.MessageCount != 3 {
		t.Fatalf("chat details = %+v; want id %v, name chat, 3 messages", details, created.ID)
	}
	if got, want := sortedIDs(details.Participants), sortedIDs([]uuid.UUID{uuid.MustParse(alice), uuid.MustParse(bob)}); !slices.Equal(got, want) {
		t.Fatalf("participants = %v; want %v", got, want)
	}

	// History is always ordered from older to newer; without a cursor it is the last page.
	historyIDs := func(params string) []int64 {
		t.Helper()
		do(t, router, http.MethodGet, chatPath+"/messages"+params, bob, "", http.StatusOK, &history)
		ids := []int64{}
		for _, m := range history.Messages {
			ids = append(ids, m.ID)
		}
		return ids
	}
	pages := []struct {
		params string
		want   []int64
	}{
		{"", []int64{first.ID, second.ID, third.ID}},
		{"?limit=2", []int64{second.ID, third.ID}},
		{fmt.Sprintf("?after=%d", first.ID), []int64{second.ID, third.ID}},
		{fmt.Sprintf("?after=%d&limit=1", first.ID), []int64{second.ID}},
		{fmt.Sprintf("?before=%d", third.ID), []int64{first.ID, second.ID}},
		{fmt.Sprintf("?before=%d&limit=1", third.ID), []int64{second.ID}},
	}
	for _, page := range pages {
		if got := historyIDs(page.params); !slices.Equal(got, page.want) {
			t.Fatalf("history %q = %v; want %v", page.params, got, page.want)
		}
	}
	do(t, router, http.MethodGet, chatPath+"/messages", bob, "", http.StatusOK, &history)
	if history.Messages[0].SenderID.String() != alice || history.Messages[0].Content != "hello" {
		t.Fatalf("first message = %+v; want hello from %s", history.Messages[0], alice)
	}
	do(t, router, http.MethodGet, chatPath+"/messages?after=1", bob, "", http.StatusBadRequest, nil)

	// Mark as read.
	lastRead := func() int64 {
		t.Helper()
		var id sql.NullInt64
		if err := db.QueryRow(`SELECT last_read_message_id FROM chat_users WHERE chat_id = $1 AND user_id = $2`, created.ID, bob).Scan(&id); err != nil {
			t.Fatalf("read mark of %s: %v", bob, err)
		}
		return id.Int64
	}
	unread := func() int {
		t.Helper()
		do(t, router, http.MethodGet, "/chats", bob, "", http.StatusOK, &chats)
		if len(chats.Chats) != 1 {
			t.Fatalf("chats = %d; want 1", len(chats.Chats))
		}
		return chats.Chats[0].UnreadCount
	}
	do(t, router, http.MethodPost, chatPath+"/read", bob, fmt.Sprintf(`{"last_read_message_id":%d}`, first.ID), http.StatusNoContent, nil)
	if lastRead() != first.ID || unread() != 2 {
		t.Fatalf("after first read: mark = %d, unread = %d; want %d and 2", lastRead(), unread(), first.ID)
	}
	do(t, router, http.MethodPost, chatPath+"/read", bob, fmt.Sprintf(`{"last_read_message_id":%d}`, third.ID), http.StatusNoContent, nil)
	if lastRead() != third.ID || unread() != 0 {
		t.Fatalf("after last read: mark = %d, unread = %d; want %d and 0", lastRead(), unread(), third.ID)
	}

	// The same or an older message is not an error and does not move the mark.
	do(t, router, http.MethodPost, chatPath+"/read", bob, fmt.Sprintf(`{"last_read_message_id":%d}`, third.ID), http.StatusNoContent, nil)
	do(t, router, http.MethodPost, chatPath+"/read", bob, fmt.Sprintf(`{"last_read_message_id":%d}`, second.ID), http.StatusNoContent, nil)
	if lastRead() != third.ID || unread() != 0 {
		t.Fatalf("after repeated and older read: mark = %d, unread = %d; want %d and 0", lastRead(), unread(), third.ID)
	}

	// A message that is not in this chat is rejected.
	do(t, router, http.MethodPost, chatPath+"/read", bob, fmt.Sprintf(`{"last_read_message_id":%d}`, foreign.ID), http.StatusBadRequest, nil)
	do(t, router, http.MethodPost, chatPath+"/read", bob, `{"last_read_message_id":1}`, http.StatusBadRequest, nil)
	if lastRead() != third.ID {
		t.Fatalf("mark after rejected read = %d; want %d", lastRead(), third.ID)
	}

	// An older value never overwrites a newer one. The domain rejects such a request before it reaches
	// the repository, so the repository is called directly, the way a concurrent request would reach it.
	messages := command.NewMessagesRepository(db)
	reader := chat.ParticipantID(uuid.MustParse(bob))
	stale := chat.SnowflakeID(first.ID)
	staleChat := chat.RehydrateChat(chat.ChatID(created.ID), chat.Name("chat"), []chat.ParticipantID{reader}, nil, map[chat.ParticipantID]*chat.SnowflakeID{reader: &stale})
	if err := messages.UpdateLastReaded(t.Context(), reader, staleChat); err != nil {
		t.Fatalf("stale read mark: %v", err)
	}
	if lastRead() != third.ID {
		t.Fatalf("mark after stale update = %d; want %d", lastRead(), third.ID)
	}

	late := message.NewMessage(message.SnowflakeID(first.ID-1), message.ChatID(created.ID), message.SenderID(uuid.MustParse(alice)), message.Content("late"))
	if _, err := messages.Create(t.Context(), late); err != nil {
		t.Fatalf("message with an older id: %v", err)
	}
	do(t, router, http.MethodGet, "/chats", bob, "", http.StatusOK, &chats)
	if c := chats.Chats[0]; c.LastMessage == nil || c.LastMessage.ID != third.ID || c.LastMessage.SenderID.String() != alice || c.LastMessage.Content != "again" {
		t.Fatalf("last message after an older one = %+v; want id %d", c.LastMessage, third.ID)
	}
	do(t, router, http.MethodGet, chatPath, bob, "", http.StatusOK, &details)
	if details.MessageCount != 4 {
		t.Fatalf("message count after an older message = %d; want 4", details.MessageCount)
	}
	if got, want := historyIDs(""), []int64{first.ID - 1, first.ID, second.ID, third.ID}; !slices.Equal(got, want) {
		t.Fatalf("history after an older message = %v; want %v", got, want)
	}
}

// sortedIDs returns the IDs as sorted strings, to compare sets of participants.
func sortedIDs(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	slices.Sort(out)
	return out
}
