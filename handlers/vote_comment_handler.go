package handlers

import (
	"encoding/json"
	"freestealer/database"
	"freestealer/models"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// VoteRequest represents a vote request
type VoteRequest struct {
	UserID   uint `json:"user_id"`
	TierID   uint `json:"tier_id"`
	VoteType int8 `json:"vote_type"` // 1 for upvote, -1 for downvote
}

// VoteTier handles POST /votes - create or update a vote
// @Summary Vote on a tier
// @Description Upvote or downvote a tier. Toggle off if same vote, change if different
// @Tags votes
// @Accept json
// @Produce json
// @Param vote body VoteRequest true "Vote data (vote_type: 1 for upvote, -1 for downvote)"
// @Success 200 {object} models.Vote
// @Success 201 {object} models.Vote
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /votes [post]
func VoteTier(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate vote type
	if req.VoteType != 1 && req.VoteType != -1 {
		http.Error(w, "Vote type must be 1 (upvote) or -1 (downvote)", http.StatusBadRequest)
		return
	}

	// Start a transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if vote already exists
	var existingVote models.Vote
	err := tx.Where("user_id = ? AND tier_id = ?", req.UserID, req.TierID).First(&existingVote).Error

	if err == gorm.ErrRecordNotFound {
		// Create new vote
		vote := models.Vote{
			UserID:   req.UserID,
			TierID:   req.TierID,
			VoteType: req.VoteType,
		}
		if err := tx.Create(&vote).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("Failed to create vote")
			http.Error(w, "Failed to create vote", http.StatusInternalServerError)
			return
		}

		// Update tier vote counts
		if req.VoteType == 1 {
			tx.Model(&models.Tier{}).Where("id = ?", req.TierID).UpdateColumn("upvote_count", gorm.Expr("upvote_count + 1"))
		} else {
			tx.Model(&models.Tier{}).Where("id = ?", req.TierID).UpdateColumn("downvote_count", gorm.Expr("downvote_count + 1"))
		}

		tx.Commit()

		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"tier_id": req.TierID,
			"type":    req.VoteType,
		}).Info("Vote created")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(vote)
		return
	}

	if err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to check existing vote")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Vote exists, update it
	oldVoteType := existingVote.VoteType

	if oldVoteType == req.VoteType {
		// Same vote, remove it (toggle off)
		if err := tx.Delete(&existingVote).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("Failed to remove vote")
			http.Error(w, "Failed to remove vote", http.StatusInternalServerError)
			return
		}

		// Update tier vote counts
		if oldVoteType == 1 {
			tx.Model(&models.Tier{}).Where("id = ?", req.TierID).UpdateColumn("upvote_count", gorm.Expr("upvote_count - 1"))
		} else {
			tx.Model(&models.Tier{}).Where("id = ?", req.TierID).UpdateColumn("downvote_count", gorm.Expr("downvote_count - 1"))
		}

		tx.Commit()

		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"tier_id": req.TierID,
		}).Info("Vote removed")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Vote removed"})
		return
	}

	// Different vote type, update it
	if err := tx.Model(&existingVote).Update("vote_type", req.VoteType).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to update vote")
		http.Error(w, "Failed to update vote", http.StatusInternalServerError)
		return
	}

	// Update tier vote counts (decrement old, increment new)
	if oldVoteType == 1 {
		tx.Model(&models.Tier{}).Where("id = ?", req.TierID).UpdateColumn("upvote_count", gorm.Expr("upvote_count - 1"))
		tx.Model(&models.Tier{}).Where("id = ?", req.TierID).UpdateColumn("downvote_count", gorm.Expr("downvote_count + 1"))
	} else {
		tx.Model(&models.Tier{}).Where("id = ?", req.TierID).UpdateColumn("downvote_count", gorm.Expr("downvote_count - 1"))
		tx.Model(&models.Tier{}).Where("id = ?", req.TierID).UpdateColumn("upvote_count", gorm.Expr("upvote_count + 1"))
	}

	tx.Commit()

	log.WithFields(log.Fields{
		"user_id":  req.UserID,
		"tier_id":  req.TierID,
		"new_type": req.VoteType,
		"old_type": oldVoteType,
	}).Info("Vote updated")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingVote)
}

// CreateComment handles POST /comments - create a new comment
// @Summary Create a comment
// @Description Add a comment to a tier (max 100 characters)
// @Tags comments
// @Accept json
// @Produce json
// @Param comment body models.Comment true "Comment data"
// @Success 201 {object} models.Comment
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /comments [post]
func CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var comment models.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate content length (max 100 characters)
	if len(comment.Content) == 0 || len(comment.Content) > 100 {
		http.Error(w, "Comment must be between 1 and 100 characters", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx := database.DB.Begin()

	// Create comment
	if err := tx.Create(&comment).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to create comment")
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// Increment comment count on tier
	if err := tx.Model(&models.Tier{}).Where("id = ?", comment.TierID).UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to update comment count")
		http.Error(w, "Failed to update comment count", http.StatusInternalServerError)
		return
	}

	tx.Commit()

	log.WithFields(log.Fields{
		"comment_id": comment.ID,
		"tier_id":    comment.TierID,
		"user_id":    comment.UserID,
	}).Info("Comment created")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// GetComments handles GET /comments?tier_id={id} - get comments for a tier
// @Summary Get comments for a tier
// @Description Get all comments for a specific tier
// @Tags comments
// @Accept json
// @Produce json
// @Param tier_id query int true "Tier ID"
// @Success 200 {array} models.Comment
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /comments [get]
func GetComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tierID := r.URL.Query().Get("tier_id")
	if tierID == "" {
		http.Error(w, "tier_id is required", http.StatusBadRequest)
		return
	}

	tid, err := strconv.ParseUint(tierID, 10, 32)
	if err != nil {
		http.Error(w, "Invalid tier_id", http.StatusBadRequest)
		return
	}

	var comments []models.Comment
	if err := database.DB.Where("tier_id = ?", tid).Preload("User").Order("created_at DESC").Find(&comments).Error; err != nil {
		log.WithError(err).Error("Failed to fetch comments")
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// DeleteComment handles DELETE /comments/{id} - delete a comment
// @Summary Delete a comment
// @Description Delete a comment from a tier
// @Tags comments
// @Accept json
// @Produce json
// @Param id path int true "Comment ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /comments/{id} [delete]
func DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// Get the comment to find tier_id before deleting
	var comment models.Comment
	if err := database.DB.First(&comment, id).Error; err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Start transaction
	tx := database.DB.Begin()

	// Delete comment
	if err := tx.Delete(&comment).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to delete comment")
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	// Decrement comment count on tier
	if err := tx.Model(&models.Tier{}).Where("id = ?", comment.TierID).UpdateColumn("comment_count", gorm.Expr("comment_count - 1")).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to update comment count")
		http.Error(w, "Failed to update comment count", http.StatusInternalServerError)
		return
	}

	tx.Commit()

	log.WithField("comment_id", id).Info("Comment deleted")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Comment deleted successfully"})
}
