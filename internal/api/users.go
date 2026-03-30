package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
	"visualds/internal/database"

	"github.com/google/uuid"
)

// DTOs

type createUserRequest struct {
	ClerkID   string `json:"clerk_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type UserReponse struct {
	UserID     uuid.UUID `json:"user_id"`
	ClerkID    string    `json:"clerk_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	CreatedAt  string    `json:"created_at"`
	UpdatedAt  string    `json:"updated_at"`
	BlockID    *string   `json:"block_id"`
	MiddleName *string   `json:"middle_name"`
	CourseID   *string   `json:"course_id"`
}

// Mappers

func ToUserResponse(u database.User) UserReponse {
	res := UserReponse{
		UserID:    u.UserID,
		ClerkID:   u.ClerkID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Time.Format(time.RFC3339),
	}

	if u.BlockID.Valid {
		val := u.BlockID.String
		res.BlockID = &val
	}

	if u.MiddleName.Valid {
		val := u.MiddleName.String
		res.MiddleName = &val
	}

	if u.CourseID.Valid {
		res.CourseID = NullToUUID(u.CourseID)
	}

	return res
}

// Handlers

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.Logger.Warn("failed to decode signup request",
			"error", err,
		)

		s.CreateErrorResponseJSON(w, "Invalid JSON format", http.StatusBadRequest)
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
		s.Logger.Warn("failed to create user",
			"error", err,
			"clerk_id", req.ClerkID,
		)
		s.CreateErrorResponseJSON(w, "User already exists", http.StatusConflict)
		return
	}

	res := ToUserResponse(user)
	s.CreateJSONResponse(w, 201, res)
}

func (s *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id, err := uuid.Parse(idString)
	if err != nil {
		s.Logger.Warn("failed to parse string UUID",
			"error", err,
			"invalid_id", idString,
		)
		s.CreateErrorResponseJSON(w, "Invalid user_id format", http.StatusBadRequest)
	}

	user, err := s.DB.GetUserByID(r.Context(), id)
	if err != nil {
		s.Logger.Warn("user not found",
			"error", err,
		)
		s.CreateErrorResponseJSON(w, "User not found", http.StatusNotFound)
		return
	}

	res := ToUserResponse(user)
	s.CreateJSONResponse(w, 200, res)
}

func (s *Server) GetAllUser(w http.ResponseWriter, r *http.Request) {
	dbUser, err := s.DB.GetAllUsers(r.Context())
	if err != nil {
		s.Logger.Error("error getting all user",
			"error", err)
		s.CreateErrorResponseJSON(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	usersDTO := make([]UserReponse, len(dbUser))
	for i, u := range dbUser {
		usersDTO[i] = ToUserResponse(u)
	}

	s.CreateJSONResponse(w, 200, usersDTO)
}
