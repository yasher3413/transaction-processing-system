package middleware

import (
	"net/http"
	"strings"
)

// APIKeyAuth middleware validates API key
func APIKeyAuth(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectedKey == "" {
				// No auth required
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("X-API-Key")
			if authHeader == "" {
				authHeader = r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					authHeader = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if authHeader != expectedKey {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Unauthorized"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}


