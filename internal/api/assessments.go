package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"visualds/internal/database"
)

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
	ID       string          `json:"id"`
	Text     string          `json:"text"`
	ImageURL *string         `json:"image_url,omitempty"`
	Type     string          `json:"type"`
	Choices  []ChoicePayload `json:"choices"`
	Feedback FeedbackPayload `json:"feedback"`
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
			Ids:          cIds,
			QuestionIds:  cQuestionIds,
			Texts:        cTexts,
			IsCorrects:   cIsCorrects,
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
	category := r.PathValue("category")
	id := r.PathValue("id")

	if category == "" || id == "" {
		s.CreateErrorResponseJSON(w, "category and id are required", http.StatusBadRequest)
		return
	}

	// 1. Get Assessment
	assessment, err := s.DB.GetAssessment(r.Context(), database.GetAssessmentParams{
		Category: category,
		ID:       id,
	})
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

	s.CreateJSONResponse(w, http.StatusOK, assessments)
}
