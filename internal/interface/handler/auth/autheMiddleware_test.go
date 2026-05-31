package auth_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/ImaSerix/go-message-service/internal/interface/handler/auth"
	"github.com/google/uuid"
)

func TestAutheMiddleware_NoHeader(t *testing.T) {
	rec, req, nextCalled, users, handler := setupTest(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := auth.UserID(r.Context()); ok {
			t.Fatal("expected userID NOT in context, but it was set")
		}
	})

	handler.ServeHTTP(rec, req)

	if *nextCalled {
		t.Fatal("expected next handler NOT to be called, but it was")
	}
	if users.called {
		t.Fatal("expected usersQuery NOT to be called, but it was")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status code %d (unauthorized), got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAutheMiddleware_InvalidUUID(t *testing.T) {
	rec, req, nextCalled, users, handler := setupTest(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := auth.UserID(r.Context()); ok {
			t.Fatal("expected userID NOT in context, but it was set")
		}
	})

	req.Header.Set("X-User-ID", "non-valid")
	handler.ServeHTTP(rec, req)

	if *nextCalled {
		t.Fatal("expected next handler NOT to be called, but it was")
	}
	if users.called {
		t.Fatal("expected usersQuery NOT to be called, but it was")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status code %d (unauthorized), got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAutheMiddleware_QueryError(t *testing.T) {
	rec, req, nextCalled, users, handler := setupTest(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := auth.UserID(r.Context()); ok {
			t.Fatal("expected userID NOT in context, but it was set")
		}
	})

	users.err = errors.New("query error")

	req.Header.Set("X-User-ID", uuid.New().String())
	handler.ServeHTTP(rec, req)

	if *nextCalled {
		t.Fatal("expected next handler NOT to be called, but it was")
	}
	if !users.called {
		t.Fatal("expected usersQuery to be called, but it was not")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status code %d (internal server error), got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestAutheMiddleware_UserNotExists(t *testing.T) {
	rec, req, nextCalled, users, handler := setupTest(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := auth.UserID(r.Context()); ok {
			t.Fatal("expected userID NOT in context, but it was set")
		}
	})

	req.Header.Set("X-User-ID", uuid.New().String())
	handler.ServeHTTP(rec, req)

	if *nextCalled {
		t.Fatal("expected next handler NOT to be called, but it was")
	}
	if !users.called {
		t.Fatal("expected usersQuery to be called, but it was not")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status code %d (unauthorized), got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAutheMiddleware_Success(t *testing.T) {

	id := uuid.New()
	rec, req, nextCalled, users, handler := setupTest(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.UserID(r.Context())
		if !ok {
			t.Fatal("expected userID in context, but it was not set")
		}
		if userID != id {
			t.Fatalf("expected userID %s, got %s", id, userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	users.exists = true

	req.Header.Set("X-User-ID", id.String())
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status code %d (ok), got %d", http.StatusOK, rec.Code)
	}
	if !*nextCalled {
		t.Fatal("expected next handler to be called, but it was not")
	}
	if !users.called {
		t.Fatal("expected usersQuery to be called, but it was not")
	}
}
