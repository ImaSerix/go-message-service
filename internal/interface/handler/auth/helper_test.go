package auth_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/ImaSerix/go-message-service/internal/interface/handler/auth"
)

func setupTest(nextFunc func(w http.ResponseWriter, r *http.Request)) (rec *httptest.ResponseRecorder, req *http.Request, nextCalled *bool, users *fakeUsersQry, handler http.Handler) {
	called := false
	nextCalled = &called

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*nextCalled = true
		nextFunc(w, r)
	})

	users = &fakeUsersQry{}
	mw := auth.AutheMiddleware(users)

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()

	return rec, req, nextCalled, users, mw(next)
}
