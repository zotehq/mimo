package middlewares

import (
	"net/http"
)

func CustomHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Server", "Krofi")

		next.ServeHTTP(w, r)
	})
}
