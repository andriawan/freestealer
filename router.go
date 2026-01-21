package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"freestealer/auth"
	"freestealer/handlers"

	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

// authMiddleware wraps handlers to require JWT authentication for protected routes
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the route is public (auth routes and health check)
		path := r.URL.Path
		publicPaths := []string{
			"/health",
			"/auth/register",
			"/auth/login",
			"/auth/github",
			"/auth/github/callback",
			"/auth/refresh",
			"/swagger/",
		}

		// Check if path starts with any public path
		for _, publicPath := range publicPaths {
			if strings.HasPrefix(path, publicPath) {
				next(w, r)
				return
			}
		}

		// For protected routes, require JWT token
		auth.RequireJWTAuth(next)(w, r)
	}
}

// SetupRoutes configures all HTTP routes for the application
func SetupRoutes(port string) {
	// Health check (public)
	http.HandleFunc("/health", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Health check called")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			log.WithError(err).Error("Failed to encode health check response")
		}
	}))

	// Authentication endpoints (public)
	http.HandleFunc("/auth/register", authMiddleware(auth.RegisterHandler))
	http.HandleFunc("/auth/login", authMiddleware(auth.LoginHandler))
	http.HandleFunc("/auth/github", authMiddleware(auth.BeginAuthHandler))
	http.HandleFunc("/auth/github/callback", authMiddleware(auth.CallbackHandler))
	http.HandleFunc("/auth/logout", authMiddleware(auth.LogoutHandler))
	http.HandleFunc("/auth/me", authMiddleware(auth.GetCurrentUser))
	http.HandleFunc("/auth/refresh", authMiddleware(auth.RefreshTokenHandler))

	// User endpoints (protected)
	http.HandleFunc("/users", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetUsers(w, r)
		case http.MethodPost:
			handlers.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Tier endpoints (protected)
	http.HandleFunc("/tiers", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTiers(w, r)
		case http.MethodPost:
			handlers.CreateTier(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/tiers/", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
	}))

	// Vote endpoint (protected)
	http.HandleFunc("/votes", authMiddleware(handlers.VoteTier))

	// Comment endpoints (protected)
	http.HandleFunc("/comments", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetComments(w, r)
		case http.MethodPost:
			handlers.CreateComment(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/comments/", authMiddleware(handlers.DeleteComment))

	// Swagger documentation (public)
	http.HandleFunc("/swagger/", authMiddleware(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	)))
}
