package api

import (
	"context"

	"github.com/google/uuid"
)

// JWT userID field name
const userIDKey string = "user_id"

func (s *Server) GetJWTUserID(ctx context.Context) (uuid.UUID, bool) {
	val, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}
	return val, true
}
