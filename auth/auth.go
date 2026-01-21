package auth

import (
	"encoding/json"
	"fmt"
	"freestealer/database"
	"freestealer/models"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	log "github.com/sirupsen/logrus"
)

var (
	// Session store for managing user sessions
	store *sessions.CookieStore
)

// InitAuth initializes the authentication system
func InitAuth() {
	// Initialize session store
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "default-secret-change-in-production"
		log.Warn("SESSION_SECRET not set, using default (not secure for production)")
	}
	store = sessions.NewCookieStore([]byte(sessionSecret))
	gothic.Store = store

	// Get GitHub OAuth credentials from environment
	githubKey := os.Getenv("GITHUB_CLIENT_ID")
	githubSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	callbackURL := os.Getenv("GITHUB_CALLBACK_URL")

	if githubKey == "" || githubSecret == "" {
		log.Fatal("GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET must be set")
	}

	if callbackURL == "" {
		callbackURL = "http://localhost:5050/auth/github/callback"
	}

	// Configure GitHub OAuth provider
	goth.UseProviders(
		github.New(githubKey, githubSecret, callbackURL),
	)

	log.Info("Authentication initialized with GitHub OAuth")
}

// BeginAuthHandler initiates GitHub OAuth flow
// @Summary Start GitHub OAuth login
// @Description Redirects user to GitHub for authentication
// @Tags auth
// @Accept json
// @Produce json
// @Success 302 {string} string "Redirect to GitHub"
// @Router /auth/github [get]
func BeginAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Set provider name in query params for gothic
	q := r.URL.Query()
	q.Add("provider", "github")
	r.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(w, r)
}

// CallbackHandler handles GitHub OAuth callback
// @Summary GitHub OAuth callback
// @Description Handles the callback from GitHub after authentication
// @Tags auth
// @Accept json
// @Produce json
// @Param code query string true "OAuth code"
// @Param state query string true "OAuth state"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/github/callback [get]
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Set provider name in query params for gothic
	q := r.URL.Query()
	if q.Get("provider") == "" {
		q.Add("provider", "github")
		r.URL.RawQuery = q.Encode()
	}

	// Complete authentication
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.WithError(err).Error("Failed to complete GitHub authentication")
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	log.WithFields(log.Fields{
		"github_id": user.UserID,
		"email":     user.Email,
		"name":      user.Name,
	}).Info("User authenticated via GitHub")

	// Find or create user in database
	var dbUser models.User
	result := database.DB.Where("github_id = ?", user.UserID).First(&dbUser)

	if result.Error != nil {
		// User doesn't exist, create new user
		username := user.NickName
		if username == "" {
			username = user.Name
		}

		dbUser = models.User{
			Username:     username,
			Email:        user.Email,
			GitHubID:     user.UserID,
			GitHubLogin:  user.NickName,
			AvatarURL:    user.AvatarURL,
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
		}

		if err := database.DB.Create(&dbUser).Error; err != nil {
			log.WithError(err).Error("Failed to create user")
			http.Error(w, "Failed to create user account", http.StatusInternalServerError)
			return
		}

		log.WithField("user_id", dbUser.ID).Info("New user created")
	} else {
		// Update existing user's tokens
		dbUser.AccessToken = user.AccessToken
		dbUser.RefreshToken = user.RefreshToken
		dbUser.AvatarURL = user.AvatarURL
		database.DB.Save(&dbUser)

		log.WithField("user_id", dbUser.ID).Info("Existing user logged in")
	}

	// Create session
	session, _ := store.Get(r, "auth-session")
	session.Values["user_id"] = dbUser.ID
	session.Values["github_id"] = user.UserID
	session.Save(r, w)

	// Return user info and redirect info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Authentication successful",
		"user": map[string]interface{}{
			"id":         dbUser.ID,
			"username":   dbUser.Username,
			"email":      dbUser.Email,
			"avatar_url": dbUser.AvatarURL,
		},
	})
}

// LogoutHandler handles user logout
// @Summary Logout user
// @Description Destroys the user session
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/logout [get]
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth-session")
	session.Values["user_id"] = nil
	session.Values["github_id"] = nil
	session.Options.MaxAge = -1
	session.Save(r, w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// GetCurrentUser returns the currently authenticated user
// @Summary Get current user
// @Description Get the currently authenticated user's information
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth-session")
	userID, ok := session.Values["user_id"].(uint)

	if !ok || userID == 0 {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// RequireAuth middleware to protect routes
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "auth-session")
		userID, ok := session.Values["user_id"].(uint)

		if !ok || userID == 0 {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Add user_id to request context or header for handlers to use
		r.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
		next(w, r)
	}
}
