package api

import (
	"context"
	"database/sql"
	"errors"
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
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	val := r.Context().Value("user_id")
	userid, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dbProgress, err := s.DB.GetAllLessonProgressByUser(ctx, userid)
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
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	val := r.Context().Value("user_id")
	userID, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rawCategory := r.PathValue("category")
	rawID := r.PathValue("id")

	lessonCategory := strings.ToLower(strings.TrimSpace(rawCategory))
	lessonID := strings.ToLower(strings.TrimSpace(rawID))

	if lessonCategory == "" || lessonID == "" {
		s.Logger.Warn("invalid lesson_category or lesson_id",
			"lesson_category", lessonCategory,
			"lesson_id", lessonID,
		)
		s.CreateErrorResponseJSON(w, "Invalid lesson parameters", http.StatusBadRequest)
		return
	}

	dbEntry, err := s.DB.CreateLessonProgressEntry(ctx, database.CreateLessonProgressEntryParams{
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

func (s *Server) DeleteLessonProgress(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	val := r.Context().Value("user_id")
	userid, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rawCategory := r.PathValue("category")
	rawID := r.PathValue("id")

	lessonCategory := strings.ToLower(strings.TrimSpace(rawCategory))
	lessonID := strings.ToLower(strings.TrimSpace(rawID))

	if lessonCategory == "" || lessonID == "" {
		s.Logger.Warn("invalid lesson_category or lesson_id",
			"lesson_category", lessonCategory,
			"lesson_id", lessonID,
		)
		s.CreateErrorResponseJSON(w, "Invalid lesson parameters", http.StatusBadRequest)
		return
	}

	progress, err := s.DB.DeleteLessonProgress(ctx, database.DeleteLessonProgressParams{
		UserID:         userid,
		LessonCategory: lessonCategory,
		LessonID:       lessonID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.Logger.Warn("lesson progress not found",
				"error", err,
				"user_id", userid,
				"lesson_category", lessonCategory,
				"lesson_id", lessonID,
			)
			s.CreateErrorResponseJSON(w, "lesson progress not found", http.StatusNotFound)
			return
		}

		s.Logger.Error("error deleting user progress entry",
			"error", err,
			"user_id", userid,
			"lesson_category", lessonCategory,
			"lesson_id", lessonID,
		)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := ToLessonProgress(progress)
	s.CreateJSONResponse(w, http.StatusOK, res)
}

func (s *Server) GetLessonProgress(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	val := r.Context().Value("user_id")
	userid, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rawCategory := r.PathValue("category")
	rawID := r.PathValue("id")

	lessonCategory := strings.ToLower(strings.TrimSpace(rawCategory))
	lessonID := strings.ToLower(strings.TrimSpace(rawID))

	if lessonCategory == "" || lessonID == "" {
		s.Logger.Warn("invalid lesson_category or lesson_id",
			"lesson_category", lessonCategory,
			"lesson_id", lessonID,
		)
		s.CreateErrorResponseJSON(w, "Invalid lesson parameters", http.StatusBadRequest)
		return
	}

	progress, err := s.DB.GetLessonProgressByID(ctx, database.GetLessonProgressByIDParams{
		UserID:         userid,
		LessonCategory: lessonCategory,
		LessonID:       lessonID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.Logger.Warn("lesson progress not found",
				"error", err,
				"user_id", userid,
				"lesson_category", lessonCategory,
				"lesson_id", lessonID,
			)
			s.CreateErrorResponseJSON(w, "lesson progress not found", http.StatusNotFound)
			return
		}

		s.Logger.Error("error deleting user progress entry",
			"error", err,
			"user_id", userid,
			"lesson_category", lessonCategory,
			"lesson_id", lessonID,
		)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := ToLessonProgress(progress)
	s.CreateJSONResponse(w, http.StatusOK, res)
}

func (s *Server) DeleteAllLessonProgress(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	val := r.Context().Value("user_id")
	userid, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := s.DB.DeleteAllLessonProgress(ctx, userid)
	if err != nil {
		s.Logger.Error("error deleting user progress",
			"error", err,
			"user_id", userid,
		)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(204)
}

func (s *Server) DeleteCategoryLessonProgress(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	val := r.Context().Value("user_id")
	userid, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rawCategory := r.PathValue("category")

	lessonCategory := strings.ToLower(strings.TrimSpace(rawCategory))

	if lessonCategory == "" {
		s.Logger.Warn("invalid lesson_category",
			"lesson_category", lessonCategory,
		)
		s.CreateErrorResponseJSON(w, "Invalid lesson parameters", http.StatusBadRequest)
		return
	}

	err := s.DB.DeleteCategoryLessonProgress(ctx, database.DeleteCategoryLessonProgressParams{
		UserID:         userid,
		LessonCategory: lessonCategory,
	})
	if err != nil {
		s.Logger.Error("error deleting user progress category",
			"error", err,
			"user_id", userid,
			"lesson_category", lessonCategory,
		)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(204)
}
