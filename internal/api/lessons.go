package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"visualds/internal/database"

	"github.com/google/uuid"
)

type CreateLessonCategoryRequest struct {
	Slug        string  `json:"slug"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	OrderIndex  int32   `json:"order_index"`
}

type UpdateLessonCategoryRequest struct {
	Slug        string  `json:"slug"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	OrderIndex  int32   `json:"order_index"`
}

type CreateLessonRequest struct {
	CategoryID uuid.UUID `json:"category_id"`
	Slug       string    `json:"slug"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	OrderIndex int32     `json:"order_index"`
}

type UpdateLessonRequest struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	OrderIndex int32  `json:"order_index"`
}

type LessonResponse struct {
	LessonID    uuid.UUID `json:"lesson_id"`
	CategoryID  uuid.UUID `json:"category_id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	OrderIndex  int32     `json:"order_index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LessonCategoryResponse struct {
	CategoryID  uuid.UUID `json:"category_id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	OrderIndex  int32     `json:"order_index"`
	LessonCount int32     `json:"lesson_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategoryWithLessonsResponse struct {
	Category LessonCategoryResponse `json:"category"`
	Lessons  []LessonResponse      `json:"lessons"`
}

func ToLessonResponse(l database.Lesson) LessonResponse {
	return LessonResponse{
		LessonID:   l.LessonID,
		CategoryID: l.CategoryID,
		Slug:       l.Slug,
		Title:      l.Title,
		Content:    l.Content,
		OrderIndex: l.OrderIndex,
		CreatedAt:  l.CreatedAt,
		UpdatedAt:  l.UpdatedAt,
	}
}

func (s *Server) GetLessonCategories(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	categories, err := s.DB.GetLessonCategories(ctx)
	if err != nil {
		s.Logger.Error("error getting lesson categories", "error", err)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := make([]LessonCategoryResponse, len(categories))
	for i, c := range categories {
		var desc *string
		if c.Description.Valid {
			desc = &c.Description.String
		}
		res[i] = LessonCategoryResponse{
			CategoryID:  c.CategoryID,
			Slug:        c.Slug,
			Title:       c.Title,
			Description: desc,
			OrderIndex:  c.OrderIndex,
			LessonCount: int32(c.LessonCount),
			CreatedAt:   c.CreatedAt,
			UpdatedAt:   c.UpdatedAt,
		}
	}

	s.CreateJSONResponse(w, http.StatusOK, res)
}

func (s *Server) GetLessonsByCategory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	categorySlug := r.PathValue("categorySlug")
	if categorySlug == "" {
		s.CreateErrorResponseJSON(w, "category slug is required", http.StatusBadRequest)
		return
	}

	// Fetch category info
	dbCategory, err := s.DB.GetLessonCategoryBySlug(ctx, categorySlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.CreateErrorResponseJSON(w, "category not found", http.StatusNotFound)
			return
		}
		s.Logger.Error("error getting category", "error", err, "slug", categorySlug)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Fetch lessons
	lessons, err := s.DB.GetLessonsByCategorySlug(ctx, categorySlug)
	if err != nil {
		s.Logger.Error("error getting lessons by category", "error", err, "slug", categorySlug)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	lessonRes := make([]LessonResponse, len(lessons))
	for i, l := range lessons {
		lessonRes[i] = ToLessonResponse(l)
	}

	var desc *string
	if dbCategory.Description.Valid {
		desc = &dbCategory.Description.String
	}

	res := CategoryWithLessonsResponse{
		Category: LessonCategoryResponse{
			CategoryID:  dbCategory.CategoryID,
			Slug:        dbCategory.Slug,
			Title:       dbCategory.Title,
			Description: desc,
			OrderIndex:  dbCategory.OrderIndex,
			LessonCount: int32(len(lessons)),
			CreatedAt:   dbCategory.CreatedAt,
			UpdatedAt:   dbCategory.UpdatedAt,
		},
		Lessons: lessonRes,
	}

	s.CreateJSONResponse(w, http.StatusOK, res)
}

func (s *Server) GetLesson(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	categorySlug := r.PathValue("categorySlug")
	lessonSlug := r.PathValue("lessonSlug")

	if categorySlug == "" || lessonSlug == "" {
		s.CreateErrorResponseJSON(w, "category and lesson slugs are required", http.StatusBadRequest)
		return
	}

	lesson, err := s.DB.GetLessonBySlug(ctx, database.GetLessonBySlugParams{
		Slug:   categorySlug,
		Slug_2: lessonSlug,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.CreateErrorResponseJSON(w, "lesson not found", http.StatusNotFound)
			return
		}
		s.Logger.Error("error getting lesson", "error", err, "category", categorySlug, "lesson", lessonSlug)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusOK, ToLessonResponse(lesson))
}

func (s *Server) GetLessonByID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	idStr := strings.TrimPrefix(r.URL.Path, "/sub-lessons/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		s.CreateErrorResponseJSON(w, "invalid lesson id", http.StatusBadRequest)
		return
	}

	lesson, err := s.DB.GetLessonByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.CreateErrorResponseJSON(w, "lesson not found", http.StatusNotFound)
			return
		}
		s.Logger.Error("error getting lesson by id", "error", err, "id", id)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusOK, ToLessonResponse(lesson))
}

// Admin handlers

func (s *Server) CreateLessonCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateLessonCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.CreateErrorResponseJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}

	category, err := s.DB.CreateLessonCategory(r.Context(), database.CreateLessonCategoryParams{
		Slug:        req.Slug,
		Title:       req.Title,
		Description: StringToNull(req.Description),
		OrderIndex:  req.OrderIndex,
	})
	if err != nil {
		s.Logger.Error("error creating lesson category", "error", err)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusCreated, category)
}

func (s *Server) UpdateLessonCategory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		s.CreateErrorResponseJSON(w, "invalid category id", http.StatusBadRequest)
		return
	}

	var req UpdateLessonCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.CreateErrorResponseJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}

	category, err := s.DB.UpdateLessonCategory(r.Context(), database.UpdateLessonCategoryParams{
		CategoryID:  id,
		Slug:        req.Slug,
		Title:       req.Title,
		Description: StringToNull(req.Description),
		OrderIndex:  req.OrderIndex,
	})
	if err != nil {
		s.Logger.Error("error updating lesson category", "error", err, "id", id)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusOK, category)
}

func (s *Server) DeleteLessonCategory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		s.CreateErrorResponseJSON(w, "invalid category id", http.StatusBadRequest)
		return
	}

	if err := s.DB.DeleteLessonCategory(r.Context(), id); err != nil {
		s.Logger.Error("error deleting lesson category", "error", err, "id", id)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) CreateLesson(w http.ResponseWriter, r *http.Request) {
	var req CreateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.CreateErrorResponseJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}

	lesson, err := s.DB.CreateLesson(r.Context(), database.CreateLessonParams{
		CategoryID: req.CategoryID,
		Slug:       req.Slug,
		Title:      req.Title,
		Content:    req.Content,
		OrderIndex: req.OrderIndex,
	})
	if err != nil {
		s.Logger.Error("error creating lesson", "error", err)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusCreated, ToLessonResponse(lesson))
}

func (s *Server) UpdateLesson(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		s.CreateErrorResponseJSON(w, "invalid lesson id", http.StatusBadRequest)
		return
	}

	var req UpdateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.CreateErrorResponseJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}

	lesson, err := s.DB.UpdateLesson(r.Context(), database.UpdateLessonParams{
		LessonID:   id,
		Title:      req.Title,
		Content:    req.Content,
		OrderIndex: req.OrderIndex,
	})
	if err != nil {
		s.Logger.Error("error updating lesson", "error", err, "id", id)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusOK, ToLessonResponse(lesson))
}

func (s *Server) DeleteLesson(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		s.CreateErrorResponseJSON(w, "invalid lesson id", http.StatusBadRequest)
		return
	}

	if err := s.DB.DeleteLesson(r.Context(), id); err != nil {
		s.Logger.Error("error deleting lesson", "error", err, "id", id)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
