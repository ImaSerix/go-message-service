package message

import (
	"net/http"
	"strconv"

	"github.com/ImaSerix/go-message-service/internal/handler"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

// TODO Скорее всего нужно уточнить, что как бы требуется вызывать middleware, которая будет как бы собирать параметры из запроса
func ParamMiddleware() func(http.Handler) http.Handler {
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
			}

			beforeParam := r.URL.Query().Get("before")
			if beforeParam != "" {
				before, err := strconv.ParseInt(beforeParam, 10, 64)
				if err != nil {
					// TODO: этот лог в целом можно удалить, просто во время продакшена мб нужен будет
					handler.LogError(r, err)
					http.Error(w, "invalid before param", http.StatusBadRequest)
					return
				}
				ctx = WithBeforeParam(ctx, before)
			}

			afterParam := r.URL.Query().Get("after")
			if afterParam != "" {
				after, err := strconv.ParseInt(afterParam, 10, 64)
				if err != nil {
					// TODO: этот лог в целом можно удалить, просто во время продакшена мб нужен будет
					handler.LogError(r, err)
					http.Error(w, "invalid after param", http.StatusBadRequest)
					return
				}
				ctx = WithAfterParam(ctx, after)
			}

			limitParam := r.URL.Query().Get("limit")
			if limitParam != "" {
				limit, err := strconv.ParseInt(limitParam, 10, strconv.IntSize)
				if err != nil {
					// TODO: этот лог в целом можно удалить, просто во время продакшена мб нужен будет
					handler.LogError(r, err)
					http.Error(w, "invalid limit param", http.StatusBadRequest)
					return
				}

				ctx = WithLimitParam(ctx, int(limit))
			} else {
				// TODO: Может быть имеет смысл такое поведение дать именно usecase, и оставить middleawre выполнять только функцию чтения
				ctx = WithLimitParam(ctx, 50)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
