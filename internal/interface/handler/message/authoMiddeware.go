package message

import (
	"errors"
	"net/http"

	"github.com/ImaSerix/go-message-service/internal/interface/handler"
	"github.com/ImaSerix/go-message-service/internal/interface/handler/auth"
)

func AuthoMiddleware(chats ChatsQry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// NOTE: Этот middleware ожидает, что было вызвоно middleware пользователей
			// TODO: Также нужно бы по хорошему добавить описание и к этому middleware
			userID, ok := auth.UserID(r.Context())
			if !ok {
				handler.LogError(r, errors.New("reading userID from context failed"))
				http.Error(w, "internal error", http.StatusUnauthorized)
				return
			}

			chatID, ok := ChatID(r.Context())
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
