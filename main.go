package main

import (
	"encoding/json"
	"freestealer/auth"
	"freestealer/database"
	"freestealer/handlers"
	"net/http"
	"os"

	_ "freestealer/docs" // Import generated docs

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Free Tier API
// @version 1.0
// @description API for tracking and sharing free tier hosting platform information
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@freetier.dev

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:5050
// @BasePath /
// @schemes http https

func main() {
	// Configure logrus
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	log.SetLevel(log.InfoLevel)

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we just log if it doesn't exist
		log.Warn("No .env file found")
	}

	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./freetier.db"
	}

	if err := database.InitDatabase(dbPath); err != nil {
		log.WithError(err).Fatal("Failed to initialize database")
	}

	// Initialize authentication
	auth.InitAuth()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.WithField("port", port).Info("Starting server")

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Health check called")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Authentication endpoints
	http.HandleFunc("/auth/github", auth.BeginAuthHandler)
	http.HandleFunc("/auth/github/callback", auth.CallbackHandler)
	http.HandleFunc("/auth/logout", auth.LogoutHandler)
	http.HandleFunc("/auth/me", auth.GetCurrentUser)

	// User endpoints
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetUsers(w, r)
		case http.MethodPost:
			handlers.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Tier endpoints
	http.HandleFunc("/tiers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTiers(w, r)
		case http.MethodPost:
			handlers.CreateTier(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/tiers/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTier(w, r)
		case http.MethodPut:
			handlers.UpdateTier(w, r)
		case http.MethodDelete:
			handlers.DeleteTier(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Vote endpoint
	http.HandleFunc("/votes", handlers.VoteTier)

	// Comment endpoints
	http.HandleFunc("/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetComments(w, r)
		case http.MethodPost:
			handlers.CreateComment(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/comments/", handlers.DeleteComment)

	// Swagger documentation
	http.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:"+port+"/swagger/doc.json"),
	))

	log.WithField("address", "http://localhost:"+port).Info("Server listening")
	log.Info("Swagger UI available at http://localhost:" + port + "/swagger/index.html")
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.WithError(err).Fatal("Server failed to start")
	}
}
