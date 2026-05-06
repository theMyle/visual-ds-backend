package api

import (
	"net/http"
	"visualds/internal/database"

	"github.com/google/uuid"
)

type PublicEligibilityStatus struct {
	Eligible    bool   `json:"eligible"`
	UserName    string `json:"userName"`
	Lessons     ProgressDetail `json:"lessons"`
	Simulators  ProgressDetail `json:"simulators"`
	Assessments ProgressDetail `json:"assessments"`
}

type ProgressDetail struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
}

func (s *Server) GetPublicUserEligibility(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.PathValue("userId")
	if userIdStr == "" {
		s.CreateErrorResponseJSON(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIdStr)
	if err != nil {
		s.CreateErrorResponseJSON(w, "Invalid User ID format", http.StatusBadRequest)
		return
	}

	// 1. Fetch User Info
	user, err := s.DB.GetUserByID(r.Context(), userID)
	if err != nil {
		s.CreateErrorResponseJSON(w, "User not found", http.StatusNotFound)
		return
	}

	userName := user.FirstName
	if user.MiddleName.Valid && user.MiddleName.String != "" {
		userName += " " + user.MiddleName.String
	}
	userName += " " + user.LastName

	// 2. Calculate Lesson Progress
	totalLessons := 0
	completedLessons := 0
	categories, err := s.DB.GetLessonCategories(r.Context())
	if err == nil {
		for _, cat := range categories {
			totalLessons += int(cat.LessonCount)
		}
		progress, err := s.DB.GetAllLessonProgressByUser(r.Context(), userID)
		if err == nil {
			// Filter progress for existing categories
			validCatSlugs := make(map[string]bool)
			for _, cat := range categories {
				validCatSlugs[cat.Slug] = true
			}
			
			uniqueLessons := make(map[string]bool)
			for _, p := range progress {
				if validCatSlugs[p.LessonCategory] {
					uniqueLessons[p.LessonID] = true
				}
			}
			completedLessons = len(uniqueLessons)
		}
	}

	// 3. Calculate Simulator Progress
	totalSimulators := 0
	completedSimulators := 0
	curriculum, err := s.DB.GetSimulatorCurriculum(r.Context())
	if err == nil {
		validPaths := make(map[string]bool)
		for _, row := range curriculum {
			if row.ChallengeID.Valid {
				totalSimulators++
				path := "/simulator/" + row.SimulatorSlug + "/" + row.ChallengeSlug.String
				validPaths[path] = true
			}
		}

		simProgress, err := s.DB.ListUserSimulatorProgress(r.Context(), userID)
		if err == nil {
			completedPaths := make(map[string]bool)
			for _, p := range simProgress {
				if p.IsCompleted && validPaths[p.Path] {
					completedPaths[p.Path] = true
				}
			}
			completedSimulators = len(completedPaths)
		}
	}

	// 4. Calculate Assessment Progress
	totalAssessments := 0
	passingAssessments := 0
	assessments, err := s.DB.ListAssessments(r.Context(), 1000)
	if err == nil {
		totalAssessments = len(assessments)
		results, err := s.DB.GetQuizResultsByUser(r.Context(), database.GetQuizResultsByUserParams{
			UserID: userID,
			Limit:  1000,
		})
		if err == nil {
			bestScores := make(map[string]float64)
			for _, r := range results {
				score := float64(r.Score) / float64(r.TotalItems) * 100
				if score > bestScores[r.QuizID] {
					bestScores[r.QuizID] = score
				}
			}
			for _, a := range assessments {
				if bestScores[a.ID] >= 75 {
					passingAssessments++
				}
			}
		}
	}

	eligible := totalLessons > 0 && completedLessons >= totalLessons &&
		totalSimulators > 0 && completedSimulators >= totalSimulators &&
		totalAssessments > 0 && passingAssessments >= totalAssessments

	s.CreateJSONResponse(w, http.StatusOK, PublicEligibilityStatus{
		Eligible: eligible,
		UserName: userName,
		Lessons: ProgressDetail{
			Total:     totalLessons,
			Completed: completedLessons,
		},
		Simulators: ProgressDetail{
			Total:     totalSimulators,
			Completed: completedSimulators,
		},
		Assessments: ProgressDetail{
			Total:     totalAssessments,
			Completed: passingAssessments,
		},
	})
}
