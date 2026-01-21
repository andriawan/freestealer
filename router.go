package main

import (
	"encoding/json"
	"net/http"

	"freestealer/auth"
	"freestealer/handlers"

	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

// SetupRoutes configures all HTTP routes for the application
func SetupRoutes(port string) {
	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Health check called")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			log.WithError(err).Error("Failed to encode health check response")
		}
	})

	// Authentication endpoints
	http.HandleFunc("/auth/register", auth.RegisterHandler)
	http.HandleFunc("/auth/login", auth.LoginHandler)
	http.HandleFunc("/auth/github", auth.BeginAuthHandler)
	http.HandleFunc("/auth/github/callback", auth.CallbackHandler)
	http.HandleFunc("/auth/logout", auth.LogoutHandler)
	http.HandleFunc("/auth/me", auth.GetCurrentUser)
	http.HandleFunc("/auth/refresh", auth.RefreshTokenHandler)

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
}
