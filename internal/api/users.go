package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"
	"visualds/internal/database"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/google/uuid"
	svix "github.com/svix/svix-webhooks/go"
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
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

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
	user, err := s.DB.CreateUser(ctx, database.CreateUserParams{
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
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	val := r.Context().Value("user_id")
	userID, ok := val.(uuid.UUID)
	if !ok {
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := s.DB.GetUserByID(ctx, userID)
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

// TODO: stream instead of chugging list
func (s *Server) GetAllUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*30)
	defer cancel()

	dbUser, err := s.DB.GetAllUsers(ctx)
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

// Clerk Webhook Event Structures

type ClerkWebhookEvent struct {
	Data      json.RawMessage `json:"data"`
	Object    string          `json:"object"`
	Type      string          `json:"type"`
	Timestamp int64           `json:"timestamp"`
}

type ClerkUserData struct {
	ID             string `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	EmailAddresses []struct {
		EmailAddress string `json:"email_address"`
	} `json:"email_addresses"`
	PublicMetadata map[string]interface{} `json:"public_metadata"`
}

// HandleClerkUserWebhook processes Clerk user webhook events
func (s *Server) HandleClerkUserWebhook(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.Logger.Warn("failed to read webhook body",
			"error", err,
		)
		s.CreateErrorResponseJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify webhook signature using Svix
	wh, err := svix.NewWebhook(s.ClerkWebhookSecret)
	if err != nil {
		s.Logger.Error("failed to create svix webhook verifier",
			"error", err,
		)
		s.CreateErrorResponseJSON(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = wh.Verify(body, r.Header)
	if err != nil {
		s.Logger.Warn("webhook signature verification failed",
			"error", err,
		)
		s.CreateErrorResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse webhook event
	var event ClerkWebhookEvent
	err = json.Unmarshal(body, &event)
	if err != nil {
		s.Logger.Warn("failed to unmarshal webhook event",
			"error", err,
		)
		s.CreateErrorResponseJSON(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Only process user.created events
	if event.Type != "user.created" {
		s.Logger.Info("ignoring webhook event",
			"type", event.Type,
		)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
		return
	}

	// Extract user data
	var userData ClerkUserData
	err = json.Unmarshal(event.Data, &userData)
	if err != nil {
		s.Logger.Warn("failed to unmarshal user data",
			"error", err,
		)
		s.CreateErrorResponseJSON(w, "Invalid user data", http.StatusBadRequest)
		return
	}

	// Get email
	email := ""
	if len(userData.EmailAddresses) > 0 {
		email = userData.EmailAddresses[0].EmailAddress
	}

	s.Logger.Info("creating user from clerk webhook",
		"clerk_id", userData.ID,
		"email", email,
	)

	// Create user in database
	user, err := s.DB.CreateUser(ctx, database.CreateUserParams{
		ClerkID:    userData.ID,
		FirstName:  userData.FirstName,
		LastName:   userData.LastName,
		Email:      email,
		CourseID:   uuid.NullUUID{Valid: false},
		MiddleName: sql.NullString{Valid: false},
		BlockID:    sql.NullString{Valid: false},
	})
	if err != nil {
		s.Logger.Error("failed to create user from webhook",
			"error", err,
			"clerk_id", userData.ID,
		)
		s.CreateErrorResponseJSON(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	s.Logger.Info("user created successfully from webhook",
		"user_id", user.UserID.String(),
		"clerk_id", userData.ID,
	)

	// Update Clerk user with database user_id in public metadata
	err = s.updateClerkUserMetadata(ctx, userData.ID, user.UserID)
	if err != nil {
		s.Logger.Error("failed to update clerk user metadata",
			"error", err,
			"clerk_id", userData.ID,
			"user_id", user.UserID.String(),
		)
		// Log but don't fail - user was created successfully. Metadata sync is secondary.
	} else {
		s.Logger.Info("clerk user metadata updated successfully",
			"clerk_id", userData.ID,
			"user_id", user.UserID.String(),
		)
	}

	res := ToUserResponse(user)
	s.CreateJSONResponse(w, 201, res)
}

func (s *Server) updateClerkUserMetadata(ctx context.Context, clerkID string, userID uuid.UUID) error {
	clerk.SetKey(s.ClerkAPIKey)

	// 1. Create the map and marshal it
	metadataMap := map[string]string{
		"db_id": userID.String(),
	}
	publicMetadataBytes, err := json.Marshal(metadataMap)
	if err != nil {
		return err
	}

	// 2. Convert to json.RawMessage
	raw := json.RawMessage(publicMetadataBytes)

	// 3. Update Clerk (passing the pointer &raw)
	_, err = user.Update(ctx, clerkID, &user.UpdateParams{
		PublicMetadata: &raw,
	})

	if err != nil {
		return err
	}

	s.Logger.Info("clerk metadata updated", "user_id", userID.String())
	return nil
}
