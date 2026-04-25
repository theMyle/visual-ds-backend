package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/google/uuid"
)

func (s *Server) CORSMiddleware(next http.Handler) http.Handler {
	allowedOrigins := s.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = map[string]bool{
			"https://visualds.vercel.app": true,
			"http://localhost:3000":       true,
			"http://127.0.0.1:3000":       true,
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

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Log the incoming request path for context
		path := r.URL.Path

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			s.Logger.Warn("Auth failed: missing or malformed header",
				"path", path,
				"header_present", authHeader != "")
			s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Log JWT verification errors (Expired, Invalid Signature, etc.)
		claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{
			Token: token,
		})
		if err != nil {
			s.Logger.Error("Auth failed: JWT verification error",
				"path", path,
				"error", err.Error())
			s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 3. Log Database lookup errors (User exists in Clerk but not in your DB)
		internalUserID, err := s.DB.GetUserByClearkID(r.Context(), claims.Subject)
		if err != nil {
			s.Logger.Warn("Auth failed: Clerk user not found in local database",
				"path", path,
				"clerk_id", claims.Subject,
				"error", err.Error())
			s.CreateErrorResponseJSON(w, "Forbidden", http.StatusForbidden)
			return
		}

		// 4. Success log (Optional - keep it 'Debug' or 'Info' for dev)
		s.Logger.Info("Auth successful",
			"path", path,
			"internal_user_id", internalUserID.UserID)

		ctx := context.WithValue(r.Context(), "user_id", internalUserID.UserID)
		ctx = context.WithValue(ctx, "clerk_claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Logger.Info("AdminOnly hit", "path", r.URL.Path)
		claims, ok := r.Context().Value("clerk_claims").(*clerk.SessionClaims)
		if !ok {
			s.Logger.Warn("Admin access denied: no claims in context")
			s.CreateErrorResponseJSON(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}

		// Check for common admin role strings using Clerk's helper method
		if !claims.HasRole("admin") && !claims.HasRole("org:admin") {
			s.Logger.Warn("Admin access denied: insufficient permissions",
				"user_id", claims.Subject)
			s.CreateErrorResponseJSON(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
