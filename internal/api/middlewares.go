package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// check if user is logged in
// TODO: integrate clerkjs
func (s *Server) MockAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockID := uuid.MustParse("0b7a92d2-c941-4629-a868-a64797ebdb5c")

		// TODO: verify clerk token

		ctx := context.WithValue(r.Context(), "user_id", mockID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
