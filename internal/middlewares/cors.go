package middlewares

import (
	"net/http"
)

type CORS struct {
	allowedOrigins map[string]struct{}
}

func NewCORSMiddleware(allowedOrigins []string) *CORS {
	allowed := make(map[string]struct{})
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}
	return &CORS{allowedOrigins: allowed}
}

func (c *CORS) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		w.Header().Add("Vary", "Origin")

		if _, ok := c.allowedOrigins[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if _, ok := c.allowedOrigins["*"]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
