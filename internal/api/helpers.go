package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *Server) CreateErrorResponseJSON(w http.ResponseWriter, msg string, code int) {
	response := ErrorResponse{
		Error: msg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
	}
}

func (s *Server) CreateJSONResponse(w http.ResponseWriter, code int, content any) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(content); err != nil {
		s.Logger.Error("error encoding json",
			"error", err)
		return
	}
}

// covert *string to NullUUID
func UUIDToNull(s *string) (uuid.NullUUID, error) {
	if *s == "" || s == nil {
		return uuid.NullUUID{Valid: false}, nil
	}

	parsed, err := uuid.Parse(*s)
	if err != nil {
		return uuid.NullUUID{Valid: false}, err
	}

	return uuid.NullUUID{UUID: parsed, Valid: true}, nil
}

// convert NullUUID to *string
func NullToUUID(nu uuid.NullUUID) *string {
	if !nu.Valid {
		return nil
	}

	s := nu.UUID.String()
	return &s
}

// convert *string to sql.NullString
func StringToNull(s *string) sql.NullString {
	if *s == "" || s == nil {
		return sql.NullString{Valid: false}
	}

	return sql.NullString{String: *s, Valid: true}
}

// convert sql.NullString to *string
func NullToString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}
