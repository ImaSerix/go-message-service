package message

import (
	"errors"
	"net/http"

	"github.com/ImaSerix/go-message-service/internal/interface/handler"
	"github.com/ImaSerix/go-message-service/internal/interface/handler/auth"
	"github.com/ImaSerix/go-message-service/internal/interface/handler/chat"
)

func AuthoMiddleware(chats ChatsQry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			userID, ok := auth.UserID(r.Context())
			if !ok {
				handler.LogError(r, errors.New("reading userID from context failed"))
				http.Error(w, "internal error", http.StatusUnauthorized)
				return
			}

			chatID, ok := chat.ChatID(r.Context())
			if !ok {
				handler.LogError(r, errors.New("reading chatID from context failed"))
				http.Error(w, "internal error", http.StatusUnauthorized)
				return
			}

			isParticipant, err := chats.UserIsParticipant(r.Context(), userID, chatID)
			if err != nil {
				handler.LogError(r, err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if !isParticipant {
				http.Error(w, "unauthorizated action", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
