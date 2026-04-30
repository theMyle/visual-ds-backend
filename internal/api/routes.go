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
	protectedMux.Handle("/admin", s.AdminOnly(http.StripPrefix("/admin", adminMux)))
	protectedMux.Handle("/admin/", s.AdminOnly(http.StripPrefix("/admin", adminMux)))

	// TODO: replace mock auth middleware with clerk
	mux.Handle("/api/", s.AuthMiddleware(http.StripPrefix("/api", protectedMux)))

	// Health check
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Webhooks (no auth required - signature verification handles security)
	mux.HandleFunc("POST /webhooks/clerk/user", s.HandleClerkUserWebhook)

	// users
	protectedMux.HandleFunc("GET /users/me", s.GetUser)

	// lessons (Public)
	mux.HandleFunc("GET /lessons", s.GetLessonCategories)
	mux.HandleFunc("GET /lessons/{categorySlug}", s.GetLessonsByCategory)
	mux.HandleFunc("GET /lessons/{categorySlug}/{lessonSlug}", s.GetLesson)

	// progress
	protectedMux.HandleFunc("GET /progress", s.GetAllLessonProgress)
	protectedMux.HandleFunc("GET /progress/{category}/{id}", s.GetLessonProgress)
	protectedMux.HandleFunc("POST /progress/{category}/{id}", s.CreateLessonProgress)
	protectedMux.HandleFunc("DELETE /progress", s.DeleteAllLessonProgress)
	protectedMux.HandleFunc("DELETE /progress/{category}", s.DeleteCategoryLessonProgress)
	protectedMux.HandleFunc("DELETE /progress/{category}/{id}", s.DeleteLessonProgress)

	// assessments
	mux.HandleFunc("GET /assessments", s.ListAssessments)
	mux.Handle("GET /assessments/{id}", s.OptionalAuthMiddleware(http.HandlerFunc(s.GetAssessment)))

	protectedMux.HandleFunc("POST /assessments/submit", s.SubmitAssessment)
	protectedMux.HandleFunc("GET /assessments/results", s.GetQuizResults)

	// simulators (Public)
	mux.HandleFunc("GET /api/simulators", s.ListSimulators)
	mux.HandleFunc("GET /api/simulators/curriculum", s.GetSimulatorCurriculum)
	mux.HandleFunc("GET /api/simulators/{simulatorSlug}/challenges/{challengeSlug}", s.GetSimulatorChallenge)

	// admin
	adminMux.HandleFunc("GET /users", s.GetAllUser)
	adminMux.HandleFunc("GET /users/{id}/progress", s.GetUserProgress)
	adminMux.HandleFunc("GET /assessments", s.ListAssessments)
	adminMux.HandleFunc("POST /assessments", s.CreateAssessment)
	adminMux.HandleFunc("GET /assessments/{id}", s.GetAssessment)
	adminMux.HandleFunc("PUT /assessments/{id}", s.UpdateAssessment)
	adminMux.HandleFunc("DELETE /assessments/{id}", s.DeleteAssessment)
	adminMux.HandleFunc("POST /assessments/{id}/questions", s.AddQuestion)
	adminMux.HandleFunc("PUT /questions/{id}", s.UpdateQuestion)
	adminMux.HandleFunc("DELETE /questions/{id}", s.DeleteQuestion)

	// admin simulators
	adminMux.HandleFunc("GET /simulators", s.ListSimulatorsAdmin)
	adminMux.HandleFunc("POST /simulators", s.CreateSimulator)
	adminMux.HandleFunc("PUT /simulators/{id}", s.UpdateSimulator)
	adminMux.HandleFunc("POST /challenges", s.CreateChallenge)
	adminMux.HandleFunc("GET /challenges/{id}", s.GetChallengeAdmin)
	adminMux.HandleFunc("PUT /challenges/{id}", s.UpdateChallenge)
	adminMux.HandleFunc("DELETE /challenges/{id}", s.DeleteChallenge)

	// admin lessons
	adminMux.HandleFunc("/lessons", s.GetLessonCategories)
	adminMux.HandleFunc("POST /lessons", s.CreateLessonCategory)
	adminMux.HandleFunc("PUT /lessons/{id}", s.UpdateLessonCategory)
	adminMux.HandleFunc("DELETE /lessons/{id}", s.DeleteLessonCategory)
	adminMux.HandleFunc("POST /sub-lessons", s.CreateLesson)
	adminMux.HandleFunc("/sub-lessons/", s.GetLessonByID)
	adminMux.HandleFunc("PUT /sub-lessons/{id}", s.UpdateLesson)
	adminMux.HandleFunc("DELETE /sub-lessons/{id}", s.DeleteLesson)

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
