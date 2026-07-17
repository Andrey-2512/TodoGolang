package middlewares

import (
	"net/http"
	"strconv"
	"strings"
)

type CORS struct {
	allowedOrigins      map[string]struct{}
	allowedMethods      []string
	allowedHeaders      []string
	allowCredentials    bool
	accessControlMaxAge int
}

func NewCORSMiddleware(allowedOrigins, allowedHeaders, allowedMethods []string, allowCredentials bool, accessControlMaxAge int) *CORS {
	allowed := make(map[string]struct{})
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}
	return &CORS{allowedOrigins: allowed, allowedMethods: allowedMethods, allowedHeaders: allowedHeaders, allowCredentials: allowCredentials, accessControlMaxAge: accessControlMaxAge}
}

func (c *CORS) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		w.Header().Add("Vary", "Origin")

		if _, ok := c.allowedOrigins[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if c.allowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.allowedHeaders, ", "))
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.allowedMethods, ", "))
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(c.accessControlMaxAge))
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
