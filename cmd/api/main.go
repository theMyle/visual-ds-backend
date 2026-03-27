package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// env variables setup
	dbURL := os.Getenv("DB_URL_IPV4")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error connecting to DB: ", err)
	}

	// db connection check
	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to DB: ", err)
	}
	log.Println("DB Connection Established")

	mux := http.NewServeMux()

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Starting Server at port ", server.Addr)
	server.ListenAndServe()
}
