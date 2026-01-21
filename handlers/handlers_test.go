package handlers

import (
	"bytes"
	"encoding/json"
	"freestealer/database"
	"freestealer/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Use PostgreSQL for testing
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("TEST_DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("TEST_DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("TEST_DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	dbname := os.Getenv("TEST_DB_NAME")
	if dbname == "" {
		dbname = "freestealer_test"
	}

	dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test - PostgreSQL not available: %v", err)
		return nil
	}

	// Clean and migrate
	db.Exec("DROP SCHEMA IF EXISTS public CASCADE")
	db.Exec("CREATE SCHEMA public")

	err = db.AutoMigrate(&models.User{}, &models.Tier{}, &models.Vote{}, &models.Comment{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	database.DB = db

	t.Run("Valid user creation", func(t *testing.T) {
		user := models.User{
			Username: "testuser",
			Email:    "test@example.com",
		}
		body, _ := json.Marshal(user)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateUser(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var response models.User
		json.NewDecoder(w.Body).Decode(&response)

		if response.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got '%s'", response.Username)
		}
	})

	t.Run("Missing required fields", func(t *testing.T) {
		user := models.User{Username: "onlyusername"}
		body, _ := json.Marshal(user)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateUser(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Invalid method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()

		CreateUser(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

func TestGetUsers(t *testing.T) {
	db := setupTestDB(t)
	database.DB = db

	// Create test users with unique GitHub IDs
	db.Create(&models.User{Username: "user1", Email: "user1@example.com", GitHubID: "gh1"})
	db.Create(&models.User{Username: "user2", Email: "user2@example.com", GitHubID: "gh2"})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	GetUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var users []models.User
	json.NewDecoder(w.Body).Decode(&users)

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestCreateTier(t *testing.T) {
	db := setupTestDB(t)
	database.DB = db

	// Create a user first
	user := models.User{Username: "tieruser", Email: "tier@example.com"}
	db.Create(&user)

	t.Run("Valid tier creation", func(t *testing.T) {
		tier := models.Tier{
			UserID:      user.ID,
			Platform:    "Railway",
			Name:        "Railway Free Tier",
			Description: "Great for testing",
			IsPublic:    true,
		}
		body, _ := json.Marshal(tier)

		req := httptest.NewRequest(http.MethodPost, "/tiers", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateTier(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var response models.Tier
		json.NewDecoder(w.Body).Decode(&response)

		if response.Platform != "Railway" {
			t.Errorf("Expected platform 'Railway', got '%s'", response.Platform)
		}
	})

	t.Run("Missing required fields", func(t *testing.T) {
		tier := models.Tier{UserID: user.ID, Platform: "Railway"}
		body, _ := json.Marshal(tier)

		req := httptest.NewRequest(http.MethodPost, "/tiers", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateTier(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestGetTiers(t *testing.T) {
	db := setupTestDB(t)
	database.DB = db

	user := models.User{Username: "tieruser", Email: "tier@example.com", GitHubID: "tieruser_gh"}
	db.Create(&user)

	// Create test tiers
	publicTier := models.Tier{
		UserID:   user.ID,
		Platform: "Railway",
		Name:     "Railway Free",
		IsPublic: true,
	}
	db.Create(&publicTier)

	privateTier := models.Tier{
		UserID:   user.ID,
		Platform: "Vercel",
		Name:     "Vercel Free",
	}
	db.Create(&privateTier)
	// Explicitly set to false after creation
	db.Model(&privateTier).Update("is_public", false)

	t.Run("Get all public tiers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tiers", nil)
		w := httptest.NewRecorder()

		GetTiers(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		data := response["data"].([]interface{})
		if len(data) != 1 {
			t.Errorf("Expected 1 public tier, got %d", len(data))
		}
	})

	t.Run("Filter by platform", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tiers?platform=Railway", nil)
		w := httptest.NewRecorder()

		GetTiers(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get user's tiers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tiers?user_id=1", nil)
		w := httptest.NewRecorder()

		GetTiers(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		data := response["data"].([]interface{})
		if len(data) != 2 {
			t.Errorf("Expected 2 tiers for user, got %d", len(data))
		}
	})
}

func TestVoteTier(t *testing.T) {
	db := setupTestDB(t)
	database.DB = db

	user := models.User{Username: "voter", Email: "voter@example.com"}
	db.Create(&user)

	tier := models.Tier{UserID: user.ID, Platform: "Railway", Name: "Test Tier"}
	db.Create(&tier)

	t.Run("Create upvote", func(t *testing.T) {
		voteReq := VoteRequest{
			UserID:   user.ID,
			TierID:   tier.ID,
			VoteType: 1,
		}
		body, _ := json.Marshal(voteReq)

		req := httptest.NewRequest(http.MethodPost, "/votes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		VoteTier(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		// Check tier upvote count
		var updatedTier models.Tier
		db.First(&updatedTier, tier.ID)
		if updatedTier.UpvoteCount != 1 {
			t.Errorf("Expected upvote count 1, got %d", updatedTier.UpvoteCount)
		}
	})

	t.Run("Toggle vote off", func(t *testing.T) {
		// Vote first time
		voteReq := VoteRequest{UserID: user.ID, TierID: tier.ID, VoteType: -1}
		body, _ := json.Marshal(voteReq)
		req := httptest.NewRequest(http.MethodPost, "/votes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		VoteTier(w, req)

		// Vote again (should toggle off)
		req = httptest.NewRequest(http.MethodPost, "/votes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		VoteTier(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Invalid vote type", func(t *testing.T) {
		voteReq := VoteRequest{UserID: user.ID, TierID: tier.ID, VoteType: 5}
		body, _ := json.Marshal(voteReq)

		req := httptest.NewRequest(http.MethodPost, "/votes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		VoteTier(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestCreateComment(t *testing.T) {
	db := setupTestDB(t)
	database.DB = db

	user := models.User{Username: "commenter", Email: "commenter@example.com"}
	db.Create(&user)

	tier := models.Tier{UserID: user.ID, Platform: "Railway", Name: "Test Tier"}
	db.Create(&tier)

	t.Run("Valid comment", func(t *testing.T) {
		comment := models.Comment{
			UserID:  user.ID,
			TierID:  tier.ID,
			Content: "Great tier!",
		}
		body, _ := json.Marshal(comment)

		req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateComment(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		// Check tier comment count
		var updatedTier models.Tier
		db.First(&updatedTier, tier.ID)
		if updatedTier.CommentCount != 1 {
			t.Errorf("Expected comment count 1, got %d", updatedTier.CommentCount)
		}
	})

	t.Run("Comment too long", func(t *testing.T) {
		longContent := "This comment is way too long and exceeds the one hundred character limit that we have set for comments on tiers!"
		comment := models.Comment{
			UserID:  user.ID,
			TierID:  tier.ID,
			Content: longContent,
		}
		body, _ := json.Marshal(comment)

		req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateComment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Empty comment", func(t *testing.T) {
		comment := models.Comment{
			UserID:  user.ID,
			TierID:  tier.ID,
			Content: "",
		}
		body, _ := json.Marshal(comment)

		req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		CreateComment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestGetComments(t *testing.T) {
	db := setupTestDB(t)
	database.DB = db

	user := models.User{Username: "commenter", Email: "commenter@example.com"}
	db.Create(&user)

	tier := models.Tier{UserID: user.ID, Platform: "Railway", Name: "Test Tier"}
	db.Create(&tier)

	// Create test comments
	db.Create(&models.Comment{UserID: user.ID, TierID: tier.ID, Content: "Comment 1"})
	db.Create(&models.Comment{UserID: user.ID, TierID: tier.ID, Content: "Comment 2"})

	t.Run("Get comments for tier", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/comments?tier_id=1", nil)
		w := httptest.NewRecorder()

		GetComments(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var comments []models.Comment
		json.NewDecoder(w.Body).Decode(&comments)

		if len(comments) != 2 {
			t.Errorf("Expected 2 comments, got %d", len(comments))
		}
	})

	t.Run("Missing tier_id parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/comments", nil)
		w := httptest.NewRecorder()

		GetComments(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}
