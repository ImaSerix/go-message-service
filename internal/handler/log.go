package handler

import (
	"log"
	"net/http"
)

func LogError(r *http.Request, err error) {
	log.Printf(
		"method=%s path=%s err=%v",
		r.Method,
		r.URL.Path,
		err,
	)
}
