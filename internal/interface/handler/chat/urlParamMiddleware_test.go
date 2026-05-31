package chat_test

import (
	"net/http"
	"testing"

	"github.com/ImaSerix/go-message-service/internal/interface/handler/chat"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func TestUrlParamMiddleware_NoURLParam(t *testing.T) {
	rec, req, nextCalled, handler := setupTest("", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := chat.ChatID(r.Context()); ok {
			t.Fatal("expected chatID NOT in context, but it was set")
		}
	})

	handler.ServeHTTP(rec, req)
	if *nextCalled {
		t.Fatal("expected next handler not to be called, but it was")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status code %d (bad request), got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUrlParamMiddleware_InvalidChatID(t *testing.T) {
	chatID := "invaliduuid"
	rec, req, nextCalled, handler := setupTest(chatID, func(w http.ResponseWriter, r *http.Request) {
		if _, ok := chat.ChatID(r.Context()); ok {
			t.Fatal("expected chatID NOT in context, but it was set")
		}
		w.WriteHeader(http.StatusOK)
	})

	r := chi.NewRouter()
	r.Post("/{chatID}", handler.ServeHTTP)

	r.ServeHTTP(rec, req)

	if *nextCalled {
		t.Fatal("expected next handler not to be called, but it was")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status code %d (bad request), got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUrlParamMiddleware_Success(t *testing.T) {
	chatID := uuid.New().String()
	rec, req, nextCalled, handler := setupTest(chatID, func(w http.ResponseWriter, r *http.Request) {
		if _, ok := chat.ChatID(r.Context()); !ok {
			t.Fatal("expected chatID in context, but it was not set")
		}
	})

	r := chi.NewRouter()
	r.Post("/{chatID}", handler.ServeHTTP)

	r.ServeHTTP(rec, req)

	if !*nextCalled {
		t.Fatal("expected next handler to be called, but it was not")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status code %d (ok), got %d", http.StatusOK, rec.Code)
	}
}
