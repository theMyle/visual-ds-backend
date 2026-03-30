package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"visualds/internal/api"
	"visualds/internal/database"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	env := os.Getenv("ENV")

	var logger slog.Handler
	var DBurl string

	if env == "production" {
		DBurl = os.Getenv("DB_URL")
		logger = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		DBurl = os.Getenv("DB_URL_IPV4")
		logger = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	db, err := sql.Open("postgres", DBurl)
	if err != nil {
		log.Fatal("Error connecting to DB: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Error connecting to DB: ", err)
	}

	log.Println("DB Connection Established")

	app := api.Server{
		DB:     database.New(db),
		Logger: slog.New(logger),
		Addr:   ":8080",
	}

	httpServer := http.Server{
		Addr:    app.Addr,
		Handler: app.Routes(),
	}

	log.Println("Starting to listen at port", app.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
