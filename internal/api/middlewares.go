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
		mockID := uuid.MustParse("e7fb7fa9-1bfb-48a2-8b6f-a3c7dae30945")

		// TODO: verify clerk token

		ctx := context.WithValue(r.Context(), "user_id", mockID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
