package api

import (
	"database/sql"
	"log/slog"
	"net/http"
	"visualds/internal/database"
)

type Server struct {
	DB                 *database.Queries
	DBRaw              *sql.DB
	Logger             *slog.Logger
	Addr               string
	ClerkWebhookSecret string
	ClerkAPIKey        string
	AllowedOrigins     map[string]bool
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	protectedMux := http.NewServeMux()
	adminMux := http.NewServeMux()

	// Auth stack: Auth -> AdminOnly -> AdminMux
	protectedMux.Handle("/admin/", s.AdminOnly(http.StripPrefix("/admin", adminMux)))

	// TODO: replace mock auth middleware with clerk
	mux.Handle("/api/", s.AuthMiddleware(http.StripPrefix("/api", protectedMux)))


	// Webhooks (no auth required - signature verification handles security)
	mux.HandleFunc("POST /webhooks/clerk/user", s.HandleClerkUserWebhook)


	// users
	protectedMux.HandleFunc("GET /users/me", s.GetUser)

	// progress
	protectedMux.HandleFunc("GET /progress", s.GetAllLessonProgress)
	protectedMux.HandleFunc("GET /progress/{category}/{id}", s.GetLessonProgress)
	protectedMux.HandleFunc("POST /progress/{category}/{id}", s.CreateLessonProgress)
	protectedMux.HandleFunc("DELETE /progress", s.DeleteAllLessonProgress)
	protectedMux.HandleFunc("DELETE /progress/{category}", s.DeleteCategoryLessonProgress)
	protectedMux.HandleFunc("DELETE /progress/{category}/{id}", s.DeleteLessonProgress)

	// assessments
	mux.HandleFunc("GET /assessments", s.ListAssessments)
	mux.HandleFunc("GET /assessments/{category}/{id}", s.GetAssessment)
	
	protectedMux.HandleFunc("POST /assessments/submit", s.SubmitAssessment)
	protectedMux.HandleFunc("GET /assessments/results", s.GetQuizResults)

	// admin
	adminMux.HandleFunc("GET /users", s.GetAllUser)
	adminMux.HandleFunc("GET /assessments", s.ListAssessments)
	adminMux.HandleFunc("POST /assessments", s.CreateAssessment)
	adminMux.HandleFunc("GET /assessments/{id}", s.GetAssessment)
	adminMux.HandleFunc("PUT /assessments/{id}", s.UpdateAssessment)
	adminMux.HandleFunc("DELETE /assessments/{id}", s.DeleteAssessment)
	adminMux.HandleFunc("DELETE /questions/{id}", s.DeleteQuestion)

	adminMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.Logger.Warn("Admin sub-route not found", "path", r.URL.Path, "method", r.Method)
		http.NotFound(w, r)
	})

	// simulator progress
	protectedMux.HandleFunc("GET /simulator-progress", s.ListUserSimulatorProgress)
	protectedMux.HandleFunc("GET /simulator-progress/{category}", s.ListUserSimulatorProgressForCategory)
	protectedMux.HandleFunc("GET /simulator-progress/{category}/{path}", s.GetSimulatorProgress)
	protectedMux.HandleFunc("POST /simulator-progress/{category}", s.UpsertSimulatorProgress)

	return s.CORSMiddleware(mux)
}
