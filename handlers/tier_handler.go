package handlers

import (
	"encoding/json"
	"freestealer/database"
	"freestealer/models"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// CreateTier handles POST /tiers - create a new tier
// @Summary Create a new tier
// @Description Create a new free tier hosting platform entry
// @Tags tiers
// @Accept json
// @Produce json
// @Param tier body models.Tier true "Tier object"
// @Success 201 {object} models.Tier
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tiers [post]
func CreateTier(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tier models.Tier
	if err := json.NewDecoder(r.Body).Decode(&tier); err != nil {
		log.WithError(err).Error("Failed to decode tier request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if tier.Platform == "" || tier.Name == "" {
		http.Error(w, "Platform and name are required", http.StatusBadRequest)
		return
	}

	// Create tier in database
	if err := database.DB.Create(&tier).Error; err != nil {
		log.WithError(err).Error("Failed to create tier")
		http.Error(w, "Failed to create tier", http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"tier_id":  tier.ID,
		"platform": tier.Platform,
	}).Info("Tier created")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tier)
}

// GetTiers handles GET /tiers - get all public tiers or user's tiers
// @Summary Get all tiers
// @Description Get list of tiers with optional filters (platform, user_id, sort)
// @Tags tiers
// @Accept json
// @Produce json
// @Param platform query string false "Filter by platform name"
// @Param user_id query int false "Filter by user ID"
// @Param sort query string false "Sort order: 'recent' or by upvotes (default)"
// @Param page query int false "Page number for pagination"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /tiers [get]
func GetTiers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := database.DB.Model(&models.Tier{})

	// Filter by platform if provided
	platform := r.URL.Query().Get("platform")
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	// Filter by user_id if provided
	userID := r.URL.Query().Get("user_id")
	if userID != "" {
		uid, _ := strconv.ParseUint(userID, 10, 32)
		query = query.Where("user_id = ?", uid)
	} else {
		// If no user_id, only show public tiers
		query = query.Where("is_public = ?", true)
	}

	// Sort by upvotes by default
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "recent" {
		query = query.Order("created_at DESC")
	} else {
		query = query.Order("upvote_count DESC, created_at DESC")
	}

	// Pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := 20
	offset := (page - 1) * pageSize

	var tiers []models.Tier
	if err := query.Preload("User").Limit(pageSize).Offset(offset).Find(&tiers).Error; err != nil {
		log.WithError(err).Error("Failed to fetch tiers")
		http.Error(w, "Failed to fetch tiers", http.StatusInternalServerError)
		return
	}

	log.WithField("count", len(tiers)).Info("Fetched tiers")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": tiers,
		"page": page,
	})
}

// GetTier handles GET /tiers/{id} - get a specific tier
// @Summary Get a tier by ID
// @Description Get detailed information about a specific tier
// @Tags tiers
// @Accept json
// @Produce json
// @Param id path int true "Tier ID"
// @Success 200 {object} models.Tier
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /tiers/{id} [get]
func GetTier(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid tier ID", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		http.Error(w, "Invalid tier ID", http.StatusBadRequest)
		return
	}

	var tier models.Tier
	if err := database.DB.Preload("User").Preload("Comments.User").First(&tier, id).Error; err != nil {
		log.WithError(err).Error("Failed to fetch tier")
		http.Error(w, "Tier not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tier)
}

// UpdateTier handles PUT /tiers/{id} - update a tier
// @Summary Update a tier
// @Description Update an existing tier information
// @Tags tiers
// @Accept json
// @Produce json
// @Param id path int true "Tier ID"
// @Param tier body models.Tier true "Tier update data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tiers/{id} [put]
func UpdateTier(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid tier ID", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		http.Error(w, "Invalid tier ID", http.StatusBadRequest)
		return
	}

	var updates models.Tier
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := database.DB.Model(&models.Tier{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		log.WithError(err).Error("Failed to update tier")
		http.Error(w, "Failed to update tier", http.StatusInternalServerError)
		return
	}

	log.WithField("tier_id", id).Info("Tier updated")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Tier updated successfully"})
}

// DeleteTier handles DELETE /tiers/{id} - delete a tier
// @Summary Delete a tier
// @Description Delete a tier from the database
// @Tags tiers
// @Accept json
// @Produce json
// @Param id path int true "Tier ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tiers/{id} [delete]
func DeleteTier(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid tier ID", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		http.Error(w, "Invalid tier ID", http.StatusBadRequest)
		return
	}

	if err := database.DB.Delete(&models.Tier{}, id).Error; err != nil {
		log.WithError(err).Error("Failed to delete tier")
		http.Error(w, "Failed to delete tier", http.StatusInternalServerError)
		return
	}

	log.WithField("tier_id", id).Info("Tier deleted")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Tier deleted successfully"})
}
