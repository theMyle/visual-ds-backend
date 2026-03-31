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
	protectedMux := http.NewServeMux()
	// AdminProtectedMux := http.NewServeMux()

	// TODO: replace mock auth middleware with clerk
	mux.Handle("/api/", s.MockAuthMiddleware(http.StripPrefix("/api", protectedMux)))

	mux.HandleFunc("GET /healthz",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("Visualds backend is live!"))
		})
	mux.HandleFunc("POST /api/users", s.CreateUser)
	mux.HandleFunc("GET /api/users", s.GetAllUser)

	// users
	protectedMux.Handle("GET /api/users/{id}",
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
	mux.HandleFunc("POST /api/quizzes/{category}/{id}", http.NotFound)
	mux.HandleFunc("DELETE /api/quizzes/{category}/{id}", http.NotFound)
	mux.HandleFunc("DELETE /api/quizzes/{category}", http.NotFound)
	mux.HandleFunc("GET /api/quizzes/{category}/{id}", http.NotFound)
	mux.HandleFunc("GET /api/quizzes/{category}", http.NotFound)

	return mux
}
