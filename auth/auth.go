package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"freestealer/database"
	"freestealer/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var (
	// Session store for managing user sessions
	store *sessions.CookieStore
	// JWT secret key
	jwtSecret []byte
	// JWT token expiration duration
	jwtExpiration = 24 * time.Hour
	// Refresh token expiration duration
	refreshExpiration = 7 * 24 * time.Hour
)

// Claims represents the JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// TokenResponse represents the response containing JWT tokens
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	GitHubID string `json:"github_id,omitempty"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

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

	// Initialize JWT secret
	jwtSecretStr := os.Getenv("JWT_SECRET")
	if jwtSecretStr == "" {
		jwtSecretStr = sessionSecret
		log.Warn("JWT_SECRET not set, using SESSION_SECRET")
	}
	jwtSecret = []byte(jwtSecretStr)

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

	log.Info("Authentication initialized with GitHub OAuth and JWT")
}

// GenerateTokens creates new JWT access and refresh tokens for a user
func GenerateTokens(user *models.User) (*TokenResponse, error) {
	now := time.Now()

	// Create access token claims
	accessClaims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "freestealer",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	// Create access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Create refresh token claims (longer expiration, minimal claims)
	refreshClaims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "freestealer",
			Subject:   fmt.Sprintf("%d", user.ID),
			ID:        fmt.Sprintf("%d-%d", user.ID, now.Unix()),
		},
	}

	// Create refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(jwtExpiration.Seconds()),
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ExtractTokenFromHeader extracts the JWT token from the Authorization header
func ExtractTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// RefreshTokenHandler handles token refresh requests
// @Summary Refresh JWT token
// @Description Get a new access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := ValidateToken(req.RefreshToken)
	if err != nil {
		log.WithError(err).Warn("Invalid refresh token")
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Get user from database
	var user models.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Generate new tokens
	tokens, err := GenerateTokens(&user)
	if err != nil {
		log.WithError(err).Error("Failed to generate tokens")
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		log.WithError(err).Error("Failed to encode token response")
	}
}

// LoginHandler handles direct login requests
// @Summary Login with email/username and password
// @Description Login with email or username and password to get JWT tokens. Password is required if user has one set.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate that at least one identifier is provided
	if req.Email == "" && req.Username == "" && req.GitHubID == "" {
		http.Error(w, "Email, username, or github_id is required", http.StatusBadRequest)
		return
	}

	// Find user by email, username, or GitHub ID
	var user models.User
	query := database.DB

	switch {
	case req.Email != "":
		query = query.Where("email = ?", req.Email)
	case req.Username != "":
		query = query.Where("username = ?", req.Username)
	case req.GitHubID != "":
		query = query.Where(&models.User{GitHubID: req.GitHubID})
	}

	if err := query.First(&user).Error; err != nil {
		log.WithFields(log.Fields{
			"email":     req.Email,
			"username":  req.Username,
			"github_id": req.GitHubID,
		}).Warn("Login failed: user not found")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check password if user has one set
	if user.Password != "" {
		if req.Password == "" {
			http.Error(w, "Password is required", http.StatusBadRequest)
			return
		}
		if !CheckPasswordHash(req.Password, user.Password) {
			log.WithField("user_id", user.ID).Warn("Login failed: invalid password")
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
	}

	// Generate JWT tokens
	tokens, err := GenerateTokens(&user)
	if err != nil {
		log.WithError(err).Error("Failed to generate tokens")
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	log.WithField("user_id", user.ID).Info("User logged in via direct login")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"avatar_url": user.AvatarURL,
		},
		"tokens": tokens,
	}); err != nil {
		log.WithError(err).Error("Failed to encode login response")
	}
}

// RegisterHandler handles user registration
// @Summary Register a new user
// @Description Register a new user with username, email, and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /auth/register [post]
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
		return
	}

	// Validate password length
	if len(req.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		http.Error(w, "User with this email or username already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		log.WithError(err).Error("Failed to hash password")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Create user
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		log.WithError(err).Error("Failed to create user")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate JWT tokens
	tokens, err := GenerateTokens(&user)
	if err != nil {
		log.WithError(err).Error("Failed to generate tokens")
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"user_id":  user.ID,
		"username": user.Username,
	}).Info("New user registered")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registration successful",
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
		"tokens": tokens,
	}); err != nil {
		log.WithError(err).Error("Failed to encode registration response")
	}
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

	// Create session (for backward compatibility)
	session, err := store.Get(r, "auth-session")
	if err != nil {
		log.WithError(err).Error("Failed to get session")
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	session.Values["user_id"] = dbUser.ID
	session.Values["github_id"] = user.UserID
	if err := session.Save(r, w); err != nil {
		log.WithError(err).Error("Failed to save session")
	}

	// Generate JWT tokens
	tokens, err := GenerateTokens(&dbUser)
	if err != nil {
		log.WithError(err).Error("Failed to generate JWT tokens")
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Return user info and JWT tokens
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Authentication successful",
		"user": map[string]interface{}{
			"id":         dbUser.ID,
			"username":   dbUser.Username,
			"email":      dbUser.Email,
			"avatar_url": dbUser.AvatarURL,
		},
		"tokens": tokens,
	}); err != nil {
		log.WithError(err).Error("Failed to encode response")
	}
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
	session, err := store.Get(r, "auth-session")
	if err != nil {
		log.WithError(err).Error("Failed to get session")
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	session.Values["user_id"] = nil
	session.Values["github_id"] = nil
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		log.WithError(err).Error("Failed to save session")
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"}); err != nil {
		log.WithError(err).Error("Failed to encode response")
	}
}

// GetCurrentUser returns the currently authenticated user
// @Summary Get current user
// @Description Get the currently authenticated user's information (supports both session and JWT)
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	var userID uint

	// Try JWT authentication first
	tokenString, err := ExtractTokenFromHeader(r)
	if err == nil {
		claims, err := ValidateToken(tokenString)
		if err == nil {
			userID = claims.UserID
		}
	}

	// Fall back to session authentication
	if userID == 0 {
		session, err := store.Get(r, "auth-session")
		if err != nil {
			log.WithError(err).Error("Failed to get session")
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}
		sessionUserID, ok := session.Values["user_id"].(uint)
		if ok {
			userID = sessionUserID
		}
	}

	if userID == 0 {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.WithError(err).Error("Failed to encode user response")
	}
}

// RequireAuth middleware to protect routes (supports both session and JWT)
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userID uint

		// Try JWT authentication first
		tokenString, err := ExtractTokenFromHeader(r)
		if err == nil {
			claims, err := ValidateToken(tokenString)
			if err == nil {
				userID = claims.UserID
			}
		}

		// Fall back to session authentication
		if userID == 0 {
			session, err := store.Get(r, "auth-session")
			if err != nil {
				log.WithError(err).Error("Failed to get session")
				http.Error(w, "Session error", http.StatusInternalServerError)
				return
			}
			sessionUserID, ok := session.Values["user_id"].(uint)
			if ok {
				userID = sessionUserID
			}
		}

		if userID == 0 {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Add user_id to request context or header for handlers to use
		r.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
		next(w, r)
	}
}

// RequireJWTAuth middleware that only accepts JWT authentication
func RequireJWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := ExtractTokenFromHeader(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user_id to request header for handlers to use
		r.Header.Set("X-User-ID", fmt.Sprintf("%d", claims.UserID))
		next(w, r)
	}
}

// SetJWTSecret allows setting the JWT secret for testing
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}
