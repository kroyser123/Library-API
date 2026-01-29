package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog/log"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Interface("error", err).
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Bytes("stack", debug.Stack()).
					Msg("Panic recovered")

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
