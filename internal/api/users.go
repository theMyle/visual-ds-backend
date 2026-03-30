package api

import (
	"time"
	"visualds/internal/database"

	"github.com/google/uuid"
)

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

func ReturnUserResponse(u database.User) UserReponse {
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
