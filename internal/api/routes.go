package api

import (
	"log/slog"
	"net/http"
	"visualds/internal/database"
)

type Server struct {
	DB     *database.Queries
	Logger *slog.Logger
	Addr   string
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("Visualds backend is live!"))
		})

	mux.HandleFunc("POST /api/users", s.CreateUser)
	mux.HandleFunc("GET /api/users/{id}", s.GetUser)
	mux.HandleFunc("GET /api/users", s.GetAllUser)

	// TODO: replace mock auth middleware with clerk
	mux.Handle("GET /api/progress",
		s.MockAuthMiddleware(
			http.HandlerFunc(s.GetAllLessonProgress)))
	mux.Handle("POST /api/progress/{lesson_category}/{lesson_id}",
		s.MockAuthMiddleware(
			http.HandlerFunc(s.CreateLessonProgress)))
	mux.HandleFunc("DELETE /api/progress/{category}/{id}", http.NotFound)

	// mux.HandleFunc("GET /api/progress/{category}/{id}", http.NotFound)
	// mux.HandleFunc("DELETE /api/progress", http.NotFound)

	mux.HandleFunc("POST /api/quizzes/{id}", http.NotFound)
	mux.HandleFunc("GET /api/quizzes/{id}", http.NotFound)
	mux.HandleFunc("GET /api/quizzes", http.NotFound)

	return mux
}
