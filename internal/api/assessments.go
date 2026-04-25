package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"visualds/internal/database"
)

type AssessmentResponse struct {
	ID       string `json:"id"`
	Category string `json:"category"`
}

type QuestionStatsPayload struct {
	Correct  int32 `json:"correct"`
	Mistakes int32 `json:"mistakes"`
}

type ChoicePayload struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
}

type FeedbackPayload struct {
	Correct   string `json:"correct"`
	Incorrect string `json:"incorrect"`
}

type QuestionPayload struct {
	ID       string               `json:"id"`
	Text     string               `json:"text"`
	ImageURL *string              `json:"image_url,omitempty"`
	Type     string               `json:"type"`
	Choices  []ChoicePayload      `json:"choices"`
	Feedback FeedbackPayload      `json:"feedback"`
	Stats    QuestionStatsPayload `json:"stats"`
}

type AssessmentPayload struct {
	ID        string            `json:"id"`
	Category  string            `json:"category"`
	Questions []QuestionPayload `json:"questions"`
}

func (s *Server) CreateAssessment(w http.ResponseWriter, r *http.Request) {
	var payload AssessmentPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.CreateErrorResponseJSON(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	// 1. Create Assessment
	_, err := s.DB.CreateAssessment(r.Context(), database.CreateAssessmentParams{
		ID:       payload.ID,
		Category: payload.Category,
	})
	if err != nil {
		s.Logger.Error("failed to create assessment", "error", err)
		s.CreateErrorResponseJSON(w, "failed to create assessment", http.StatusInternalServerError)
		return
	}

	// 2. Prepare data for BulkCreateQuestions
	var qIds, qAssessmentIds, qTexts, qImageUrls, qTypes, qFeedbacksCorrect, qFeedbacksIncorrect []string

	// 3. Prepare data for BulkCreateChoices
	var cIds, cQuestionIds, cTexts []string
	var cIsCorrects []bool

	for _, q := range payload.Questions {
		// Auto-generate ID if the user didn't provide one
		if q.ID == "" {
			q.ID = uuid.New().String()
		}

		qIds = append(qIds, q.ID)
		qAssessmentIds = append(qAssessmentIds, payload.ID)
		qTexts = append(qTexts, q.Text)

		imgUrl := ""
		if q.ImageURL != nil {
			imgUrl = *q.ImageURL
		}
		qImageUrls = append(qImageUrls, imgUrl)

		qTypes = append(qTypes, q.Type)
		qFeedbacksCorrect = append(qFeedbacksCorrect, q.Feedback.Correct)
		qFeedbacksIncorrect = append(qFeedbacksIncorrect, q.Feedback.Incorrect)

		for _, c := range q.Choices {
			if c.ID == "" {
				c.ID = uuid.New().String()
			}
			cIds = append(cIds, c.ID)
			cQuestionIds = append(cQuestionIds, q.ID)
			cTexts = append(cTexts, c.Text)
			cIsCorrects = append(cIsCorrects, c.IsCorrect)
		}
	}

	// Insert Questions
	if len(qIds) > 0 {
		err = s.DB.BulkCreateQuestions(r.Context(), database.BulkCreateQuestionsParams{
			Ids:                qIds,
			AssessmentIds:      qAssessmentIds,
			Texts:              qTexts,
			ImageUrls:          qImageUrls,
			Types:              qTypes,
			FeedbacksCorrect:   qFeedbacksCorrect,
			FeedbacksIncorrect: qFeedbacksIncorrect,
		})
		if err != nil {
			s.Logger.Error("failed to create questions", "error", err)
			s.CreateErrorResponseJSON(w, "failed to create questions", http.StatusInternalServerError)
			return
		}
	}

	// Insert Choices
	if len(cIds) > 0 {
		err = s.DB.BulkCreateChoices(r.Context(), database.BulkCreateChoicesParams{
			Ids:         cIds,
			QuestionIds: cQuestionIds,
			Texts:       cTexts,
			IsCorrects:  cIsCorrects,
		})
		if err != nil {
			s.Logger.Error("failed to create choices", "error", err)
			s.CreateErrorResponseJSON(w, "failed to create choices", http.StatusInternalServerError)
			return
		}
	}

	s.CreateJSONResponse(w, http.StatusCreated, map[string]string{
		"message": "assessment created successfully",
	})
}

func (s *Server) GetAssessment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		s.CreateErrorResponseJSON(w, "id is required", http.StatusBadRequest)
		return
	}

	// 1. Get Assessment by ID only (we need a new query or modify existing)
	// For now, let's look for any category since ID is primary key
	// I'll update GetAssessment query to only use ID if category is not provided
	assessment, err := s.DB.GetAssessmentById(r.Context(), id)
	if err != nil {
		s.CreateErrorResponseJSON(w, "assessment not found", http.StatusNotFound)
		return
	}

	// 2. Get Questions
	questions, err := s.DB.GetQuestionsByAssessmentId(r.Context(), assessment.ID)
	if err != nil {
		s.CreateErrorResponseJSON(w, "failed to get questions", http.StatusInternalServerError)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		var limit int
		for _, c := range limitStr {
			if c >= '0' && c <= '9' {
				limit = limit*10 + int(c-'0')
			}
		}
		if limit > 0 && limit < len(questions) {
			questions = questions[:limit]
		}
	}

	// 3. Get all choices for these questions
	var questionIds []string
	for _, q := range questions {
		questionIds = append(questionIds, q.ID)
	}

	var choices []database.Choice
	if len(questionIds) > 0 {
		choices, err = s.DB.GetChoicesByQuestionIds(r.Context(), questionIds)
		if err != nil {
			s.CreateErrorResponseJSON(w, "failed to get choices", http.StatusInternalServerError)
			return
		}
	}

	// Organize choices by question_id
	choicesByQuestion := make(map[string][]ChoicePayload)
	for _, c := range choices {
		choicesByQuestion[c.QuestionID] = append(choicesByQuestion[c.QuestionID], ChoicePayload{
			ID:        c.ID,
			Text:      c.Text,
			IsCorrect: c.IsCorrect,
		})
	}

	// Assemble QuestionPayloads
	var questionPayloads []QuestionPayload
	for _, q := range questions {
		imgUrl := NullToString(q.ImageUrl)

		qChoices := choicesByQuestion[q.ID]
		if qChoices == nil {
			qChoices = []ChoicePayload{}
		}

		questionPayloads = append(questionPayloads, QuestionPayload{
			ID:       q.ID,
			Text:     q.Text,
			ImageURL: imgUrl,
			Type:     q.Type,
			Feedback: FeedbackPayload{
				Correct:   q.FeedbackCorrect,
				Incorrect: q.FeedbackIncorrect,
			},
			Choices: qChoices,
			Stats: QuestionStatsPayload{
				Correct:  q.CorrectCount,
				Mistakes: q.MistakeCount,
			},
		})
	}

	// Build final payload
	payload := AssessmentPayload{
		ID:        assessment.ID,
		Category:  assessment.Category,
		Questions: questionPayloads,
	}

	s.CreateJSONResponse(w, http.StatusOK, payload)
}

func (s *Server) ListAssessments(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := int32(10) // default limit

	if limitStr != "" {
		// simple ascii conversion
		var l int32
		for _, c := range limitStr {
			if c >= '0' && c <= '9' {
				l = l*10 + int32(c-'0')
			}
		}
		if l > 0 && l <= 100 {
			limit = l
		}
	}

	assessments, err := s.DB.ListAssessments(r.Context(), limit)
	if err != nil {
		s.CreateErrorResponseJSON(w, "failed to list assessments", http.StatusInternalServerError)
		return
	}

	response := make([]AssessmentResponse, len(assessments))
	for i, a := range assessments {
		response[i] = AssessmentResponse{
			ID:       a.ID,
			Category: a.Category,
		}
	}

	s.CreateJSONResponse(w, http.StatusOK, response)
}

func (s *Server) DeleteAssessment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.Logger.Info("DeleteAssessment hit", "id", id)
	if id == "" {
		s.CreateErrorResponseJSON(w, "id is required", http.StatusBadRequest)
		return
	}

	err := s.DB.DeleteAssessment(r.Context(), id)
	if err != nil {
		s.Logger.Error("failed to delete assessment", "error", err)
		s.CreateErrorResponseJSON(w, "failed to delete assessment", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusOK, map[string]string{
		"message": "assessment deleted successfully",
	})
}

func (s *Server) UpdateAssessment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		s.CreateErrorResponseJSON(w, "id is required", http.StatusBadRequest)
		return
	}

	var payload struct {
		Category string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.CreateErrorResponseJSON(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	assessment, err := s.DB.UpdateAssessment(r.Context(), database.UpdateAssessmentParams{
		ID:       id,
		Category: payload.Category,
	})
	if err != nil {
		s.Logger.Error("failed to update assessment", "error", err)
		s.CreateErrorResponseJSON(w, "failed to update assessment", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusOK, AssessmentResponse{
		ID:       assessment.ID,
		Category: assessment.Category,
	})
}

func (s *Server) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		s.CreateErrorResponseJSON(w, "id is required", http.StatusBadRequest)
		return
	}

	err := s.DB.DeleteQuestion(r.Context(), id)
	if err != nil {
		s.Logger.Error("failed to delete question", "error", err)
		s.CreateErrorResponseJSON(w, "failed to delete question", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusOK, map[string]string{
		"message": "question deleted successfully",
	})
}

func (s *Server) AddQuestion(w http.ResponseWriter, r *http.Request) {
	assessmentID := r.PathValue("id")
	if assessmentID == "" {
		s.CreateErrorResponseJSON(w, "assessment id is required", http.StatusBadRequest)
		return
	}

	var payload QuestionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.CreateErrorResponseJSON(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := s.DBRaw.BeginTx(r.Context(), nil)
	if err != nil {
		s.Logger.Error("failed to start transaction", "error", err)
		s.CreateErrorResponseJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	qtx := s.DB.WithTx(tx)

	// 1. Create Question
	if payload.ID == "" {
		payload.ID = uuid.New().String()
	}

	_, err = qtx.CreateQuestion(r.Context(), database.CreateQuestionParams{
		ID:                payload.ID,
		AssessmentID:      assessmentID,
		Text:              payload.Text,
		ImageUrl:          StringToNull(payload.ImageURL),
		Type:              payload.Type,
		FeedbackCorrect:   payload.Feedback.Correct,
		FeedbackIncorrect: payload.Feedback.Incorrect,
	})
	if err != nil {
		s.Logger.Error("failed to create question", "error", err)
		s.CreateErrorResponseJSON(w, "failed to create question", http.StatusInternalServerError)
		return
	}

	// 2. Create Choices
	for _, c := range payload.Choices {
		if c.ID == "" {
			c.ID = uuid.New().String()
		}
		_, err = qtx.CreateChoice(r.Context(), database.CreateChoiceParams{
			ID:         c.ID,
			QuestionID: payload.ID,
			Text:       c.Text,
			IsCorrect:  c.IsCorrect,
		})
		if err != nil {
			s.Logger.Error("failed to create choice", "error", err)
			s.CreateErrorResponseJSON(w, "failed to create choice", http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.Logger.Error("failed to commit transaction", "error", err)
		s.CreateErrorResponseJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusCreated, map[string]string{
		"message": "question added successfully",
		"id":      payload.ID,
	})
}

func (s *Server) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	questionID := r.PathValue("id")
	if questionID == "" {
		s.CreateErrorResponseJSON(w, "question id is required", http.StatusBadRequest)
		return
	}

	var payload QuestionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.CreateErrorResponseJSON(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := s.DBRaw.BeginTx(r.Context(), nil)
	if err != nil {
		s.Logger.Error("failed to start transaction", "error", err)
		s.CreateErrorResponseJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	qtx := s.DB.WithTx(tx)

	// 1. Update Question
	_, err = qtx.UpdateQuestion(r.Context(), database.UpdateQuestionParams{
		ID:                questionID,
		Text:              payload.Text,
		ImageUrl:          StringToNull(payload.ImageURL),
		Type:              payload.Type,
		FeedbackCorrect:   payload.Feedback.Correct,
		FeedbackIncorrect: payload.Feedback.Incorrect,
	})
	if err != nil {
		s.Logger.Error("failed to update question", "error", err)
		s.CreateErrorResponseJSON(w, "failed to update question", http.StatusInternalServerError)
		return
	}

	// 2. Replace Choices
	// Delete existing
	err = qtx.DeleteChoicesByQuestionId(r.Context(), questionID)
	if err != nil {
		s.Logger.Error("failed to delete existing choices", "error", err)
		s.CreateErrorResponseJSON(w, "failed to update choices", http.StatusInternalServerError)
		return
	}

	// Insert new
	for _, c := range payload.Choices {
		if c.ID == "" {
			c.ID = uuid.New().String()
		}
		_, err = qtx.CreateChoice(r.Context(), database.CreateChoiceParams{
			ID:         c.ID,
			QuestionID: questionID,
			Text:       c.Text,
			IsCorrect:  c.IsCorrect,
		})
		if err != nil {
			s.Logger.Error("failed to create choice", "error", err)
			s.CreateErrorResponseJSON(w, "failed to create choices", http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.Logger.Error("failed to commit transaction", "error", err)
		s.CreateErrorResponseJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusOK, map[string]string{
		"message": "question updated successfully",
	})
}
