package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

func (s *Server) CORSMiddleware(next http.Handler) http.Handler {
	allowedOrigins := s.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = map[string]bool{
			"https://visualds.vercel.app": true,
			"http://localhost:3000":      true,
			"http://127.0.0.1:3000":      true,
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// check if user is logged in
// TODO: integrate clerkjs
func (s *Server) MockAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockID := uuid.MustParse("e7fb7fa9-1bfb-48a2-8b6f-a3c7dae30945")

		// TODO: verify clerk token

		ctx := context.WithValue(r.Context(), "user_id", mockID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
