package chat_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/ImaSerix/go-message-service/internal/interface/handler/chat"
)

func setupTest(chatID string, nextFunc func(http.ResponseWriter, *http.Request)) (rec *httptest.ResponseRecorder, req *http.Request, nextCalled *bool, handler http.Handler) {
	called := false
	nextCalled = &called

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*nextCalled = true
		nextFunc(w, r)
	})
	mw := chat.UrlParamMiddleware()
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	if chatID != "" {
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/%s", chatID), nil)
	}

	rec = httptest.NewRecorder()

	return rec, req, nextCalled, mw(next)
}
