package api

import (
	"encoding/json"
	"net/http"
	"visualds/internal/database"

	"github.com/google/uuid"
)

type SimulatorChallengeResponse struct {
	ID               string          `json:"id"`
	SimulatorID      string          `json:"simulator_id"`
	Slug             string          `json:"slug"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	OrderIndex       int32           `json:"order_index"`
	InitialCode      string          `json:"initial_code"`
	ProgramStructure json.RawMessage `json:"program_structure"`
	TestCases        json.RawMessage `json:"test_cases"`
	Capacity         json.RawMessage `json:"capacity"`
	NextChallengeID  *string         `json:"next_challenge_id,omitempty"`
	NextChallengeSlug *string        `json:"next_challenge_slug,omitempty"`
}

type SimulatorCurriculumResponse struct {
	ID          string               `json:"id"`
	Slug        string               `json:"slug"`
	Name        string               `json:"name"`
	Challenges  []CurriculumChallenge `json:"challenges"`
}

type CurriculumChallenge struct {
	ID         string `json:"id"`
	Slug       string `json:"slug"`
	Title      string `json:"title"`
	OrderIndex int32  `json:"order_index"`
	Path       string `json:"path"`
}

func (s *Server) ListSimulators(w http.ResponseWriter, r *http.Request) {
	sims, err := s.DB.ListSimulators(r.Context())
	if err != nil {
		s.Logger.Error("failed to list simulators", "error", err)
		s.CreateErrorResponseJSON(w, "failed to load simulators", http.StatusInternalServerError)
		return
	}
	s.CreateJSONResponse(w, http.StatusOK, sims)
}

// ListSimulatorsAdmin returns simulators with their challenges for the admin dashboard.
func (s *Server) ListSimulatorsAdmin(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.GetSimulatorCurriculum(r.Context())
	if err != nil {
		s.Logger.Error("failed to get simulator curriculum for admin", "error", err)
		s.CreateErrorResponseJSON(w, "failed to load simulators", http.StatusInternalServerError)
		return
	}

	simMap := make(map[string]*SimulatorCurriculumResponse)
	var orderedSlugs []string

	for _, row := range rows {
		if _, ok := simMap[row.SimulatorSlug]; !ok {
			simMap[row.SimulatorSlug] = &SimulatorCurriculumResponse{
				ID:         row.SimulatorID,
				Slug:       row.SimulatorSlug,
				Name:       row.SimulatorName,
				Challenges: []CurriculumChallenge{},
			}
			orderedSlugs = append(orderedSlugs, row.SimulatorSlug)
		}

		if row.ChallengeSlug.Valid {
			simMap[row.SimulatorSlug].Challenges = append(simMap[row.SimulatorSlug].Challenges, CurriculumChallenge{
				ID:         row.ChallengeID.String,
				Slug:       row.ChallengeSlug.String,
				Title:      row.ChallengeTitle.String,
				OrderIndex: row.OrderIndex.Int32,
				Path:       "/simulator/" + row.SimulatorSlug + "/" + row.ChallengeSlug.String,
			})
		}
	}

	response := make([]*SimulatorCurriculumResponse, 0, len(orderedSlugs))
	for _, slug := range orderedSlugs {
		response = append(response, simMap[slug])
	}

	s.CreateJSONResponse(w, http.StatusOK, response)
}


func (s *Server) GetSimulatorCurriculum(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.GetSimulatorCurriculum(r.Context())
	if err != nil {
		s.Logger.Error("failed to get simulator curriculum", "error", err)
		s.CreateErrorResponseJSON(w, "failed to load curriculum", http.StatusInternalServerError)
		return
	}

	// Group by simulator
	simMap := make(map[string]*SimulatorCurriculumResponse)
	var orderedSlugs []string

	for _, row := range rows {
		if _, ok := simMap[row.SimulatorSlug]; !ok {
			simMap[row.SimulatorSlug] = &SimulatorCurriculumResponse{
				ID:         row.SimulatorID,
				Slug:       row.SimulatorSlug,
				Name:       row.SimulatorName,
				Challenges: []CurriculumChallenge{},
			}
			orderedSlugs = append(orderedSlugs, row.SimulatorSlug)
		}

		if row.ChallengeSlug.Valid {
			simMap[row.SimulatorSlug].Challenges = append(simMap[row.SimulatorSlug].Challenges, CurriculumChallenge{
				ID:         row.ChallengeID.String,
				Slug:       row.ChallengeSlug.String,
				Title:      row.ChallengeTitle.String,
				OrderIndex: row.OrderIndex.Int32,
				Path:       "/simulator/" + row.SimulatorSlug + "/" + row.ChallengeSlug.String,
			})
		}
	}

	response := make([]*SimulatorCurriculumResponse, 0, len(orderedSlugs))
	for _, slug := range orderedSlugs {
		response = append(response, simMap[slug])
	}

	s.CreateJSONResponse(w, http.StatusOK, response)
}

func (s *Server) GetSimulatorChallenge(w http.ResponseWriter, r *http.Request) {
	simulatorSlug := r.PathValue("simulatorSlug")
	challengeSlug := r.PathValue("challengeSlug")

	if simulatorSlug == "" || challengeSlug == "" {
		s.CreateErrorResponseJSON(w, "missing slug", http.StatusBadRequest)
		return
	}

	// 1. Get Simulator
	sim, err := s.DB.GetSimulatorBySlug(r.Context(), simulatorSlug)
	if err != nil {
		s.CreateErrorResponseJSON(w, "simulator not found", http.StatusNotFound)
		return
	}

	// 2. Get Challenge
	challenge, err := s.DB.GetChallengeBySlug(r.Context(), database.GetChallengeBySlugParams{
		SimulatorID: sim.ID,
		Slug:        challengeSlug,
	})
	if err != nil {
		s.CreateErrorResponseJSON(w, "challenge not found", http.StatusNotFound)
		return
	}

	resp := SimulatorChallengeResponse{
		ID:               challenge.ID,
		SimulatorID:      challenge.SimulatorID,
		Slug:             challenge.Slug,
		Title:            challenge.Title,
		Description:      challenge.Description,
		OrderIndex:       challenge.OrderIndex,
		InitialCode:      challenge.InitialCode, // This is already the COALESCE result from the query
		ProgramStructure: challenge.ProgramStructure,
		TestCases:        challenge.TestCases,
		Capacity:         challenge.Capacity,
	}

	if challenge.NextChallengeID.Valid {
		resp.NextChallengeID = &challenge.NextChallengeID.String
	}

	// If no explicit next challenge, look for the next in order
	if resp.NextChallengeID == nil {
		nextSlug, err := s.DB.GetNextChallengeByOrder(r.Context(), database.GetNextChallengeByOrderParams{
			SimulatorID: sim.ID,
			OrderIndex:  challenge.OrderIndex,
		})
		if err == nil {
			resp.NextChallengeSlug = &nextSlug
		}
	} else if challenge.NextChallengeSlug.Valid {
		resp.NextChallengeSlug = &challenge.NextChallengeSlug.String
	}

	s.CreateJSONResponse(w, http.StatusOK, resp)
}

// Admin / Internal Seeder Handlers

func (s *Server) CreateSimulator(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ID          string `json:"id"`
		Slug        string `json:"slug"`
		Name        string `json:"name"`
		Description string `json:"description"`
		InitialCode string `json:"initial_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.CreateErrorResponseJSON(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if payload.ID == "" {
		payload.ID = uuid.New().String()
	}

	sim, err := s.DB.CreateSimulator(r.Context(), database.CreateSimulatorParams{
		ID:          payload.ID,
		Slug:        payload.Slug,
		Name:        payload.Name,
		Description: payload.Description,
		InitialCode: payload.InitialCode,
		IsActive:    true,
	})
	if err != nil {
		s.CreateErrorResponseJSON(w, "failed to create simulator", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusCreated, sim)
}

func (s *Server) CreateChallenge(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ID               string          `json:"id"`
		SimulatorID      string          `json:"simulator_id"`
		Slug             string          `json:"slug"`
		Title            string          `json:"title"`
		Description      string          `json:"description"`
		OrderIndex       int32           `json:"order_index"`
		InitialCode      *string         `json:"initial_code"`
		ProgramStructure json.RawMessage `json:"program_structure"`
		TestCases        json.RawMessage `json:"test_cases"`
		Capacity         json.RawMessage `json:"capacity"`
		NextChallengeID  *string         `json:"next_challenge_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.CreateErrorResponseJSON(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if payload.ID == "" {
		payload.ID = uuid.New().String()
	}

	nextID := StringToNull(payload.NextChallengeID)

	challenge, err := s.DB.CreateChallenge(r.Context(), database.CreateChallengeParams{
		ID:               payload.ID,
		SimulatorID:      payload.SimulatorID,
		Slug:             payload.Slug,
		Title:            payload.Title,
		Description:      payload.Description,
		OrderIndex:       payload.OrderIndex,
		InitialCode:      StringToNull(payload.InitialCode),
		ProgramStructure: payload.ProgramStructure,
		TestCases:        payload.TestCases,
		Capacity:         payload.Capacity,
		NextChallengeID:  nextID,
		IsActive:         true,
	})
	if err != nil {
		s.Logger.Error("failed to create challenge", "error", err)
		s.CreateErrorResponseJSON(w, "failed to create challenge", http.StatusInternalServerError)
		return
	}

	s.CreateJSONResponse(w, http.StatusCreated, challenge)
}
