package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"visualds/internal/database"

	"github.com/google/uuid"
)

type createQuizResultRequest struct {
	Score      int32 `json:"score"`
	TotalItems int32 `json:"total_items"`
}

type QuizResultResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	QuizCategory string    `json:"quiz_category"`
	QuizID       string    `json:"quiz_id"`
	Score        int32     `json:"score"`
	TotalItems   int32     `json:"total_items"`
	TakenAt      time.Time `json:"taken_at"`
}

func ToQuizResultResponse(qr database.QuizResult) QuizResultResponse {
	return QuizResultResponse{
		ID:           qr.ID,
		UserID:       qr.UserID,
		QuizCategory: qr.QuizCategory,
		QuizID:       qr.QuizID,
		Score:        qr.Score,
		TotalItems:   qr.TotalItems,
		TakenAt:      qr.TakenAt,
	}
}

func (s *Server) CreateQuizResult(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	userID, ok := s.GetJWTUserID(ctx)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	quizCategory := r.PathValue("category")
	quizID := r.PathValue("id")

	if quizCategory == "" || quizID == "" {
		s.Logger.Warn("empty quiz category or id when creating quiz result")
		s.CreateErrorResponseJSON(w, "invalid quiz category or ID", http.StatusBadRequest)
		return
	}

	req := createQuizResultRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.Logger.Error("error decoding json", "error", err)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	dbResult, err := s.DB.CreateQuizResultEntry(ctx, database.CreateQuizResultEntryParams{
		UserID:       userID,
		QuizCategory: quizCategory,
		QuizID:       quizID,
		Score:        req.Score,
		TotalItems:   req.TotalItems,
	})
	if err != nil {
		s.Logger.Error("error creating user quiz result entry", "error", err)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := ToQuizResultResponse(dbResult)
	s.CreateJSONResponse(w, http.StatusCreated, res)
}
