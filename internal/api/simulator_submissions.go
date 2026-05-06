package api

import (
	"context"
	"net/http"
	"time"
	"visualds/internal/database"

	"github.com/google/uuid"
)

type SimulatorSubmissionResponse struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	SimulatorID string     `json:"simulator_id"`
	ChallengeID string     `json:"challenge_id"`
	Code        string     `json:"code"`
	Status      string     `json:"status"`
	CreatedAt   *time.Time `json:"created_at"`
}

func ToSimulatorSubmission(s database.SimulatorSubmission) SimulatorSubmissionResponse {
	var createdAt *time.Time
	if s.CreatedAt.Valid {
		createdAt = &s.CreatedAt.Time
	}

	return SimulatorSubmissionResponse{
		ID:          s.ID,
		UserID:      s.UserID,
		SimulatorID: s.SimulatorID,
		ChallengeID: s.ChallengeID,
		Code:        s.Code,
		Status:      s.Status,
		CreatedAt:   createdAt,
	}
}

func (s *Server) ListUserSubmissionsForChallenge(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	val := r.Context().Value("user_id")
	userid, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	challengeID := r.PathValue("challengeId")
	if challengeID == "" {
		s.CreateErrorResponseJSON(w, "Missing challengeId", http.StatusBadRequest)
		return
	}

	dbSubmissions, err := s.DB.ListUserSubmissionsForChallenge(ctx, database.ListUserSubmissionsForChallengeParams{
		UserID:      userid,
		ChallengeID: challengeID,
	})
	if err != nil {
		s.Logger.Error("error getting simulator submissions",
			"error", err,
			"user_id", userid,
			"challenge_id", challengeID,
		)
		s.CreateErrorResponseJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := make([]SimulatorSubmissionResponse, len(dbSubmissions))
	for i, sub := range dbSubmissions {
		res[i] = ToSimulatorSubmission(sub)
	}

	s.CreateJSONResponse(w, http.StatusOK, res)
}
