package handlers

import (
	"encoding/json"
	"net/http"

	"freestealer/database"
	"freestealer/models"

	log "github.com/sirupsen/logrus"
)

// CreateUser handles POST /users - create a new user
// @Summary Create a new user
// @Description Register a new user with username and email
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.User true "User object"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Security BearerAuth
// @Router /users [post]
func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if user.Username == "" || user.Email == "" {
		http.Error(w, "Username and email are required", http.StatusBadRequest)
		return
	}

	// Create user
	if err := database.DB.Create(&user).Error; err != nil {
		log.WithError(err).Error("Failed to create user")
		http.Error(w, "Failed to create user (username or email may already exist)", http.StatusConflict)
		return
	}

	log.WithFields(log.Fields{
		"user_id":  user.ID,
		"username": user.Username,
	}).Info("User created")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.WithError(err).Error("Failed to encode user response")
	}
}

// GetUsers handles GET /users - get all users
// @Summary Get all users
// @Description Get a list of all registered users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /users [get]
func GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		log.WithError(err).Error("Failed to fetch users")
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.WithError(err).Error("Failed to encode users response")
	}
}
