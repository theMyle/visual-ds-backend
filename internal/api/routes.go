package api

import (
	"log/slog"
	"net/http"
	"visualds/internal/database"
)

type Server struct {
	DB                 *database.Queries
	Logger             *slog.Logger
	Addr               string
	ClerkWebhookSecret string
	ClerkAPIKey        string
	AllowedOrigins     map[string]bool
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	protectedMux := http.NewServeMux()

	// TODO: replace mock auth middleware with clerk
	mux.Handle("/api/", s.AuthMiddleware(http.StripPrefix("/api", protectedMux)))

	mux.HandleFunc("GET /healthz",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("Visualds backend is live!"))
		})

	// Webhooks (no auth required - signature verification handles security)
	mux.HandleFunc("POST /webhooks/clerk/user", s.HandleClerkUserWebhook)

	// mux.HandleFunc("GET /api/users", s.GetAllUser) // TODO: this should be admin only

	// users
	protectedMux.Handle("GET /api/users/me",
		s.MockAuthMiddleware(http.HandlerFunc(s.GetUser)))

	// progress
	protectedMux.HandleFunc("GET /progress", s.GetAllLessonProgress)
	protectedMux.HandleFunc("GET /progress/{category}/{id}", s.GetLessonProgress)
	protectedMux.HandleFunc("POST /progress/{category}/{id}", s.CreateLessonProgress)
	protectedMux.HandleFunc("DELETE /progress", s.DeleteAllLessonProgress)
	protectedMux.HandleFunc("DELETE /progress/{category}", s.DeleteCategoryLessonProgress)
	protectedMux.HandleFunc("DELETE /progress/{category}/{id}", s.DeleteLessonProgress)

	// quiz
	// TODO: implement routes
	protectedMux.HandleFunc("POST /quizzes/{category}/{id}", s.CreateQuizResult)

	mux.HandleFunc("GET /api/quizzes/{category}/{id}", http.NotFound)
	mux.HandleFunc("GET /api/quizzes/{category}", http.NotFound)
	mux.HandleFunc("GET /api/quizzes", http.NotFound)

	// protectedMux.HandleFunc("DELETE /api/quizzes/{category}/{id}", http.NotFound)
	// mux.HandleFunc("DELETE /api/quizzes/{category}", http.NotFound)

	return s.CORSMiddleware(mux)
}
