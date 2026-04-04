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

type QuizResultSnapshot struct {
	ID         uuid.UUID `json:"id"`
	Score      int32     `json:"score"`
	TotalItems int32     `json:"total_items"`
	TakenAt    time.Time `json:"taken_at"`
}

type QuizResultSummaryResponse struct {
	QuizCategory string             `json:"quiz_category"`
	QuizID       string             `json:"quiz_id"`
	Highest      QuizResultSnapshot `json:"highest"`
	MostRecent   QuizResultSnapshot `json:"most_recent"`
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

func toQuizResultSnapshot(qr database.QuizResult) QuizResultSnapshot {
	return QuizResultSnapshot{
		ID:         qr.ID,
		Score:      qr.Score,
		TotalItems: qr.TotalItems,
		TakenAt:    qr.TakenAt,
	}
}

func (s *Server) GetAllQuizResults(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	userID, ok := s.GetJWTUserID(ctx)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dbSummaries, err := s.DB.GetQuizResultSummariesByUser(ctx, userID)
	if err != nil {
		s.Logger.Error("error getting quiz results", "error", err, "user_id", userID)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := make([]QuizResultSummaryResponse, 0, len(dbSummaries))
	for _, summary := range dbSummaries {
		res = append(res, QuizResultSummaryResponse{
			QuizCategory: summary.QuizCategory,
			QuizID:       summary.QuizID,
			Highest: QuizResultSnapshot{
				ID:         summary.HighestID,
				Score:      summary.HighestScore,
				TotalItems: summary.HighestTotalItems,
				TakenAt:    summary.HighestTakenAt,
			},
			MostRecent: QuizResultSnapshot{
				ID:         summary.MostRecentID,
				Score:      summary.MostRecentScore,
				TotalItems: summary.MostRecentTotalItems,
				TakenAt:    summary.MostRecentTakenAt,
			},
		})
	}

	s.CreateJSONResponse(w, http.StatusOK, res)
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
