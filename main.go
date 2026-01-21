package main

import (
	"net/http"
	"os"
	"time"

	_ "freestealer/docs" // Import generated docs

	"freestealer/auth"
	"freestealer/database"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
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

	// Initialize database (PostgreSQL)
	if err := database.InitDatabase(); err != nil {
		log.WithError(err).Fatal("Failed to initialize database")
	}

	// Initialize authentication
	auth.InitAuth()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.WithField("port", port).Info("Starting server")

	// Setup all routes
	SetupRoutes(port)

	log.WithField("address", "http://localhost:"+port).Info("Server listening")
	log.Info("Swagger UI available at http://localhost:" + port + "/swagger/index.html")

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.WithError(err).Fatal("Server failed to start")
	}
}
