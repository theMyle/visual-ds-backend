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

// DTOs

type SimulatorProgressResponse struct {
	UserID            uuid.UUID  `json:"user_id"`
	SimulatorCategory string     `json:"simulator_category"`
	Path              string     `json:"path"`
	IsCompleted       bool       `json:"is_completed"`
	UpdatedAt         *time.Time `json:"updated_at"`
}

type UpsertSimulatorProgressRequest struct {
	Path        string `json:"path"`
	IsCompleted bool   `json:"is_completed"`
}

// Mappers

func ToSimulatorProgress(s database.SimulatorProgress) SimulatorProgressResponse {
	var updatedAt *time.Time
	if s.UpdatedAt.Valid {
		updatedAt = &s.UpdatedAt.Time
	}

	return SimulatorProgressResponse{
		UserID:            s.UserID,
		SimulatorCategory: s.SimulatorCategory,
		Path:              s.Path,
		IsCompleted:       s.IsCompleted,
		UpdatedAt:         updatedAt,
	}
}

// Handlers

func (s *Server) ListUserSimulatorProgress(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	val := r.Context().Value("user_id")
	userid, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dbProgress, err := s.DB.ListUserSimulatorProgress(ctx, userid)
	if err != nil {
		s.Logger.Warn("error getting simulator progress",
			"error", err,
			"user_id", userid,
		)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := make([]SimulatorProgressResponse, len(dbProgress))
	for i, l := range dbProgress {
		res[i] = ToSimulatorProgress(l)
	}

	s.CreateJSONResponse(w, 200, res)
}

func (s *Server) ListUserSimulatorProgressForCategory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	val := r.Context().Value("user_id")
	userid, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rawCategory := r.PathValue("category")
	simulatorCategory := strings.ToLower(strings.TrimSpace(rawCategory))

	if simulatorCategory == "" {
		s.Logger.Warn("invalid simulator_category",
			"simulator_category", simulatorCategory,
		)
		s.CreateErrorResponseJSON(w, "Invalid simulator parameters", http.StatusBadRequest)
		return
	}

	dbProgress, err := s.DB.ListUserSimulatorProgressForCategory(ctx, database.ListUserSimulatorProgressForCategoryParams{
		UserID:            userid,
		SimulatorCategory: simulatorCategory,
	})
	if err != nil {
		s.Logger.Warn("error getting simulator progress for category",
			"error", err,
			"user_id", userid,
			"simulator_category", simulatorCategory,
		)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := make([]SimulatorProgressResponse, len(dbProgress))
	for i, l := range dbProgress {
		res[i] = ToSimulatorProgress(l)
	}

	s.CreateJSONResponse(w, 200, res)
}

func (s *Server) GetSimulatorProgress(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	val := r.Context().Value("user_id")
	userid, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rawCategory := r.PathValue("category")
	rawPath := r.PathValue("path")

	simulatorCategory := strings.ToLower(strings.TrimSpace(rawCategory))
	path := strings.ToLower(strings.TrimSpace(rawPath))

	if simulatorCategory == "" || path == "" {
		s.Logger.Warn("invalid simulator_category or path",
			"simulator_category", simulatorCategory,
			"path", path,
		)
		s.CreateErrorResponseJSON(w, "Invalid simulator parameters", http.StatusBadRequest)
		return
	}

	progress, err := s.DB.GetSimulatorProgress(ctx, database.GetSimulatorProgressParams{
		UserID: userid,
		Path:   path,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.Logger.Warn("simulator progress not found",
				"error", err,
				"user_id", userid,
				"path", path,
			)
			s.CreateErrorResponseJSON(w, "simulator progress not found", http.StatusNotFound)
			return
		}

		s.Logger.Error("error getting simulator progress entry",
			"error", err,
			"user_id", userid,
			"path", path,
		)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if progress.SimulatorCategory != simulatorCategory {
		s.CreateErrorResponseJSON(w, "simulator progress category mismatch", http.StatusNotFound)
		return
	}

	res := ToSimulatorProgress(progress)
	s.CreateJSONResponse(w, http.StatusOK, res)
}

func (s *Server) UpsertSimulatorProgress(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	val := r.Context().Value("user_id")
	userID, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rawCategory := r.PathValue("category")
	simulatorCategory := strings.ToLower(strings.TrimSpace(rawCategory))

	if simulatorCategory == "" {
		s.Logger.Warn("invalid simulator_category",
			"simulator_category", simulatorCategory,
		)
		s.CreateErrorResponseJSON(w, "Invalid simulator parameters", http.StatusBadRequest)
		return
	}

	var req UpsertSimulatorProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Logger.Warn("invalid json body",
			"error", err,
		)
		s.CreateErrorResponseJSON(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	path := strings.ToLower(strings.TrimSpace(req.Path))
	if path == "" {
		s.Logger.Warn("invalid path in json body")
		s.CreateErrorResponseJSON(w, "Path is required", http.StatusBadRequest)
		return
	}

	err := s.DB.UpsertSimulatorProgress(ctx, database.UpsertSimulatorProgressParams{
		UserID:            userID,
		SimulatorCategory: simulatorCategory,
		Path:              path,
		IsCompleted:       req.IsCompleted,
	})
	if err != nil {
		s.Logger.Error("error upserting simulator progress entry",
			"error", err)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	progress, err := s.DB.GetSimulatorProgress(ctx, database.GetSimulatorProgressParams{
		UserID: userID,
		Path:   path,
	})
	if err != nil {
		s.Logger.Error("error fetching upserted simulator progress entry",
			"error", err)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := ToSimulatorProgress(progress)
	s.CreateJSONResponse(w, http.StatusCreated, res)
}
