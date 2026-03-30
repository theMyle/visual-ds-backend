package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"visualds/internal/database"

	"github.com/google/uuid"
)

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	type signUpRequest struct {
		ClerkID   string `json:"clerk_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	}

	var req signUpRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("Request error:", err)
		CreateErrorResponseJSON(w, "Invalid or wrong request", http.StatusBadRequest)
		return
	}

	// Insert to DB
	user, err := s.DB.CreateUser(r.Context(), database.CreateUserParams{
		ClerkID:    req.ClerkID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		CourseID:   uuid.NullUUID{Valid: false},
		MiddleName: sql.NullString{Valid: false},
		BlockID:    sql.NullString{Valid: false},
	})
	if err != nil {
		log.Println("DB error:", err)
		CreateErrorResponseJSON(w, "User ID already exists or database error", http.StatusInternalServerError)
		return
	}

	resp := ReturnUserResponse(user)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println(err)
		return
	}
}
