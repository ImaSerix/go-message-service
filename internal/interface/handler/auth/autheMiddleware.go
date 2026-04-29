package auth

import (
	"net/http"
	"strings"

	"github.com/ImaSerix/go-message-service/internal/interface/handler"
	"github.com/google/uuid"
)

func AutheMiddleware(users UsersQry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			userIDHeader := r.Header.Get("X-User-ID")

			if strings.TrimSpace(userIDHeader) == "" {
				http.Error(w, "missing userID", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(userIDHeader)
			if err != nil {
				http.Error(w, "invalid userID: cannot parse id", http.StatusUnauthorized)
				return
			}

			exists, err := users.Exists(r.Context(), userID)
			if err != nil {
				handler.LogError(r, err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if !exists {
				http.Error(w, "invalid userID: no such user localy", http.StatusUnauthorized)
				return
			}

			ctx := WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
