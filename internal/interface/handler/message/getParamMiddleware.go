package message

import (
	"net/http"
	"strconv"
)

func GetParamMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			beforeParam := r.URL.Query().Get("before")
			if beforeParam != "" {
				before, err := strconv.ParseInt(beforeParam, 10, 64)
				if err != nil {
					http.Error(w, "invalid before param", http.StatusBadRequest)
					return
				}
				ctx = WithBeforeParam(ctx, before)
			}

			afterParam := r.URL.Query().Get("after")
			if afterParam != "" {
				after, err := strconv.ParseInt(afterParam, 10, 64)
				if err != nil {
					http.Error(w, "invalid after param", http.StatusBadRequest)
					return
				}
				ctx = WithAfterParam(ctx, after)
			}

			limitParam := r.URL.Query().Get("limit")
			if limitParam != "" {
				limit, err := strconv.ParseInt(limitParam, 10, strconv.IntSize)
				if err != nil || limit < 1 {
					http.Error(w, "invalid limit param", http.StatusBadRequest)
					return
				}

				ctx = WithLimitParam(ctx, int(limit))
			} else {
				ctx = WithLimitParam(ctx, 50)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
