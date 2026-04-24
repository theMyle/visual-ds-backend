package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"visualds/internal/database"

	"github.com/google/uuid"
)

type QuestionOutcome struct {
	QuestionID string `json:"question_id"`
	IsCorrect  bool   `json:"is_correct"`
}

type SubmitAssessmentRequest struct {
	QuizID       string            `json:"quiz_id"`
	QuizCategory string            `json:"quiz_category"`
	Score        int               `json:"score"`
	TotalItems   int               `json:"total_items"`
	Outcomes     []QuestionOutcome `json:"outcomes"`
}

func (s *Server) SubmitAssessment(w http.ResponseWriter, r *http.Request) {
	// 1. Get user from context
	val := r.Context().Value("user_id")
	userID, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Parse payload
	var req SubmitAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.CreateErrorResponseJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Start transaction
	tx, err := s.DBRaw.BeginTx(r.Context(), nil)
	if err != nil {
		s.Logger.Error("failed to start transaction", "error", err)
		s.CreateErrorResponseJSON(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	qtx := s.DB.WithTx(tx)

	// 4. Save session summary
	result, err := qtx.SaveQuizResult(r.Context(), database.SaveQuizResultParams{
		UserID:       userID,
		QuizCategory: req.QuizCategory,
		QuizID:       req.QuizID,
		Score:        int32(req.Score),
		TotalItems:   int32(req.TotalItems),
	})
	if err != nil {
		s.Logger.Error("failed to save quiz result", "error", err)
		s.CreateErrorResponseJSON(w, "Failed to save result", http.StatusInternalServerError)
		return
	}

	// 5. Update global question stats
	for _, outcome := range req.Outcomes {
		var correct, mistakes int32
		if outcome.IsCorrect {
			correct = 1
		} else {
			mistakes = 1
		}

		err := qtx.UpdateQuestionStats(r.Context(), database.UpdateQuestionStatsParams{
			QuestionID: outcome.QuestionID,
			Correct:    correct,
			Mistakes:   mistakes,
		})
		if err != nil {
			s.Logger.Error("failed to update question stats", "question_id", outcome.QuestionID, "error", err)
			// We continue even if stats fail, or should we fail the whole thing?
			// Since stats are for content optimization, failing the whole submission might be too aggressive.
			// However, in a transaction, if we return error, it rollbacks.
			// Let's stick to atomicity.
			s.CreateErrorResponseJSON(w, "Failed to update stats", http.StatusInternalServerError)
			return
		}
	}

	// 6. Commit transaction
	if err := tx.Commit(); err != nil {
		s.Logger.Error("failed to commit transaction", "error", err)
		s.CreateErrorResponseJSON(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusCreated, result)
}

func (s *Server) GetQuizResults(w http.ResponseWriter, r *http.Request) {
	// 1. Get user from context
	val := r.Context().Value("user_id")
	userID, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Parse limit from query params
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // Sane default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 3. Fetch from DB
	results, err := s.DB.GetQuizResultsByUser(r.Context(), database.GetQuizResultsByUserParams{
		UserID: userID,
		Limit:  int32(limit),
	})
	if err != nil {
		s.Logger.Error("failed to fetch quiz results", "error", err, "user_id", userID)
		s.CreateErrorResponseJSON(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusOK, results)
}
