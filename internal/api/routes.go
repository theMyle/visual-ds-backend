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
	protectedMux.HandleFunc("GET /users/me", s.GetUser)

	// progress
	protectedMux.HandleFunc("GET /progress", s.GetAllLessonProgress)
	protectedMux.HandleFunc("GET /progress/{category}/{id}", s.GetLessonProgress)
	protectedMux.HandleFunc("POST /progress/{category}/{id}", s.CreateLessonProgress)
	protectedMux.HandleFunc("DELETE /progress", s.DeleteAllLessonProgress)
	protectedMux.HandleFunc("DELETE /progress/{category}", s.DeleteCategoryLessonProgress)
	protectedMux.HandleFunc("DELETE /progress/{category}/{id}", s.DeleteLessonProgress)


	// simulator progress
	protectedMux.HandleFunc("GET /simulator-progress", s.ListUserSimulatorProgress)
	protectedMux.HandleFunc("GET /simulator-progress/{category}", s.ListUserSimulatorProgressForCategory)
	protectedMux.HandleFunc("GET /simulator-progress/{category}/{path}", s.GetSimulatorProgress)
	protectedMux.HandleFunc("POST /simulator-progress/{category}", s.UpsertSimulatorProgress)

	return s.CORSMiddleware(mux)
}
