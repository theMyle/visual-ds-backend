package api

import (
	"net/http"
	"visualds/internal/database"
)

type Server struct {
	DB   *database.Queries
	Addr string
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/users", s.CreateUser)
	mux.HandleFunc("PUT /api/users/{id}", http.NotFound)
	mux.HandleFunc("GET /api/users/{id}", http.NotFound)
	mux.HandleFunc("GET /api/users", http.NotFound)

	return mux
}
