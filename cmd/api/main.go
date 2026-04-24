package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"visualds/internal/api"
	"visualds/internal/database"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment variables")
	}

	env := os.Getenv("ENV")
	port := os.Getenv("PORT")

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

	if len(port) > 0 {
		port = ":" + port
	} else {
		port = ":8080"
	}

	db, err := sql.Open("postgres", DBurl)
	if err != nil {
		log.Fatal("Error connecting to DB: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Error connecting to DB: ", err)
	}

	log.Println("DB Connection Established")

	// Load Clerk environment variables
	clerkAPIKey := os.Getenv("CLERK_API_KEY")
	clerkWebhookSecret := os.Getenv("CLERK_WEBHOOK_SECRET")
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")

	clerk.SetKey(clerkAPIKey)

	allowedOrigins := map[string]bool{}
	if allowedOriginsEnv != "" {
		for _, origin := range strings.Split(allowedOriginsEnv, ",") {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				allowedOrigins[trimmed] = true
			}
		}
	}

	if clerkWebhookSecret == "" {
		log.Println("WARNING: CLERK_WEBHOOK_SECRET not set; webhook signature verification will not work correctly")
	}

	app := api.Server{
		DB:                 database.New(db),
		DBRaw:              db,
		Logger:             slog.New(logger),
		Addr:               port,
		ClerkAPIKey:        clerkAPIKey,
		ClerkWebhookSecret: clerkWebhookSecret,
		AllowedOrigins:     allowedOrigins,
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
