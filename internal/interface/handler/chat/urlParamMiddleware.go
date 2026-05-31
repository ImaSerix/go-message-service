package chat

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func UrlParamMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			chatIDParam := chi.URLParam(r, "chatID")
			if chatIDParam != "" {
				chatID, ok := uuid.Parse(chatIDParam)
				if ok != nil {
					http.Error(w, "invalid chatID param", http.StatusBadRequest)
					return
				}
				ctx = WithChatID(ctx, chatID)
			} else {
				http.Error(w, "invalid chatID param", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
