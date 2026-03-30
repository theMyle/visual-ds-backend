package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"visualds/internal/api"
	"visualds/internal/database"

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

	app := api.Server{
		DB:   database.New(db),
		Addr: ":8080",
	}

	httpServer := http.Server{
		Addr:    app.Addr,
		Handler: app.Routes(),
	}

	log.Println("Start listening at port", app.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
