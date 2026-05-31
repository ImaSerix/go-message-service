package message_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ImaSerix/go-message-service/internal/interface/handler/message"
)

const defaultLimit = 50

func TestGetParamMiddleware(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		expBefore int64
		expAfter  int64
		expLimit  int
		beforeSet bool
		afterSet  bool
		limitSet  bool
		expCalled bool
		expCode   int
	}{
		{
			name:      "invalid before",
			url:       "/?before=ab",
			expCalled: false,
			expCode:   http.StatusBadRequest,
		},
		{
			name:      "invalid after",
			url:       "/?after=ab",
			expCalled: false,
			expCode:   http.StatusBadRequest,
		},
		{
			name:      "invalid limit",
			url:       "/?limit=ab",
			expCalled: false,
			expCode:   http.StatusBadRequest,
		},
		{
			name:      "valid before",
			url:       "/?before=50",
			beforeSet: true,
			expBefore: 50,
			expCalled: true,
			expCode:   http.StatusOK,
		},
		{
			name:      "valid after",
			url:       "/?after=50",
			afterSet:  true,
			expAfter:  50,
			expCalled: true,
			expCode:   http.StatusOK,
		},
		{
			name:      "valid limit",
			url:       "/?limit=100",
			limitSet:  true,
			expLimit:  100,
			expCalled: true,
			expCode:   http.StatusOK,
		},
		{
			name:      "all set",
			url:       "/?before=40&after=20&limit=100",
			beforeSet: true,
			expBefore: 40,
			afterSet:  true,
			expAfter:  20,
			limitSet:  true,
			expLimit:  100,
			expCalled: true,
			expCode:   http.StatusOK,
		},
		{
			name:      "nothing set",
			url:       "/",
			expCalled: true,
			expCode:   http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				before, ok := message.BeforeParam(r.Context())
				if ok != test.beforeSet {
					if test.beforeSet {
						t.Fatal("expected before in context, but it was not set")
					} else {
						t.Fatal("expected before not in context, but it was set")
					}
				}
				if ok && before != test.expBefore {
					t.Fatalf("expected before %d, got %d", test.expBefore, before)
				}

				after, ok := message.AfterParam(r.Context())
				if ok != test.afterSet {
					if test.afterSet {
						t.Fatal("expected after in context, but it was not set")
					} else {
						t.Fatal("expected after not in context, but it was set")
					}
				}
				if ok && after != test.expAfter {
					t.Fatalf("expected after %d, got %d", test.expAfter, after)
				}

				limit, ok := message.LimitParam(r.Context())
				if !ok {
					t.Fatal("expected limit in context always is set, but it is not")
				}
				if !test.limitSet && limit != defaultLimit {
					t.Fatalf("expected limit %d, got %d", defaultLimit, limit)
				}
				if test.limitSet && limit != test.expLimit {
					t.Fatalf("expected limit %d, got %d", test.expLimit, limit)
				}
				w.WriteHeader(http.StatusOK)
			})

			mw := message.GetParamMiddleware()
			req := httptest.NewRequest(http.MethodPost, test.url, nil)
			rec := httptest.NewRecorder()

			mw(next).ServeHTTP(rec, req)

			if called != test.expCalled {
				t.Fatalf("expected next called = %v, got %v", test.expCalled, called)
			}
			if rec.Code != test.expCode {
				t.Fatalf("expected status code %d, got %d", test.expCode, rec.Code)
			}
		})
	}
}
