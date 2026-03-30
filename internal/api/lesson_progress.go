package api

import (
	"net/http"
	"strings"
	"time"
	"visualds/internal/database"

	"github.com/google/uuid"
)

// DTOs

type LessonProgressResponse struct {
	UserID         uuid.UUID `json:"user_id"`
	LessonCategory string    `json:"lesson_category"`
	LessonID       string    `json:"lesson_id"`
	CompletedAt    time.Time `json:"completed_at"`
}

// Mappers

func ToLessonProgress(l database.LessonProgress) LessonProgressResponse {
	return LessonProgressResponse{
		UserID:         l.UserID,
		LessonCategory: l.LessonCategory,
		LessonID:       l.LessonID,
		CompletedAt:    l.CompletedAt,
	}
}

// Handlers

func (s *Server) GetAllLessonProgress(w http.ResponseWriter, r *http.Request) {
	// i need userid here
	val := r.Context().Value("user_id")
	userid, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dbProgress, err := s.DB.GetAllLessonProgressByUser(r.Context(), userid)
	if err != nil {
		s.Logger.Warn("error getting lesson progress",
			"error", err,
			"user_id", userid,
		)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := make([]LessonProgressResponse, len(dbProgress))
	for i, l := range dbProgress {
		res[i] = ToLessonProgress(l)
	}

	s.CreateJSONResponse(w, 200, res)
}

func (s *Server) CreateLessonProgress(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value("user_id")
	userID, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rawCategory := r.PathValue("lesson_category")
	rawID := r.PathValue("lesson_id")

	lessonCategory := strings.ToLower(strings.TrimSpace(rawCategory))
	lessonID := strings.ToLower(strings.TrimSpace(rawID))

	if lessonCategory == "" || lessonID == "" {
		s.Logger.Warn("invalid lesson_category or lesson_id", "lesson_category", lessonCategory,
			"lesson_id", lessonID)
		s.CreateErrorResponseJSON(w, "Invalid lesson parameters", http.StatusBadRequest)
		return
	}

	dbEntry, err := s.DB.CreateLessonProgressEntry(r.Context(), database.CreateLessonProgressEntryParams{
		UserID:         userID,
		LessonCategory: lessonCategory,
		LessonID:       lessonID,
	})
	if err != nil {
		s.Logger.Error("error creating progress entry",
			"error", err)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
	}

	res := ToLessonProgress(dbEntry)
	s.CreateJSONResponse(w, http.StatusCreated, res)
}
