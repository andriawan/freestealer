package models

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(&User{}, &Tier{}, &Vote{}, &Comment{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestUserModel(t *testing.T) {
	db := setupTestDB(t)

	t.Run("Create user", func(t *testing.T) {
		user := User{
			Username:    "testuser",
			Email:       "test@example.com",
			GitHubID:    "12345",
			GitHubLogin: "testuser",
		}

		result := db.Create(&user)
		if result.Error != nil {
			t.Errorf("Failed to create user: %v", result.Error)
		}

		if user.ID == 0 {
			t.Error("User ID should be set after creation")
		}
	})

	t.Run("Unique username constraint", func(t *testing.T) {
		user1 := User{Username: "unique", Email: "user1@example.com"}
		db.Create(&user1)

		user2 := User{Username: "unique", Email: "user2@example.com"}
		result := db.Create(&user2)

		if result.Error == nil {
			t.Error("Expected error for duplicate username")
		}
	})

	t.Run("Unique email constraint", func(t *testing.T) {
		user1 := User{Username: "user1", Email: "same@example.com"}
		db.Create(&user1)

		user2 := User{Username: "user2", Email: "same@example.com"}
		result := db.Create(&user2)

		if result.Error == nil {
			t.Error("Expected error for duplicate email")
		}
	})
}

func TestTierModel(t *testing.T) {
	db := setupTestDB(t)

	// Create a user first
	user := User{Username: "tieruser", Email: "tier@example.com"}
	db.Create(&user)

	t.Run("Create tier", func(t *testing.T) {
		tier := Tier{
			UserID:      user.ID,
			Platform:    "Railway",
			Name:        "Railway Free Tier",
			Description: "Great for small projects",
			IsPublic:    true,
			CPULimit:    "0.5 vCPU",
			MemoryLimit: "512MB",
		}

		result := db.Create(&tier)
		if result.Error != nil {
			t.Errorf("Failed to create tier: %v", result.Error)
		}

		if tier.ID == 0 {
			t.Error("Tier ID should be set after creation")
		}
	})

	t.Run("Default values", func(t *testing.T) {
		tier := Tier{
			UserID:   user.ID,
			Platform: "Vercel",
			Name:     "Vercel Free",
		}
		db.Create(&tier)

		if tier.UpvoteCount != 0 {
			t.Errorf("Expected default upvote count 0, got %d", tier.UpvoteCount)
		}
		if tier.DownvoteCount != 0 {
			t.Errorf("Expected default downvote count 0, got %d", tier.DownvoteCount)
		}
		if tier.CommentCount != 0 {
			t.Errorf("Expected default comment count 0, got %d", tier.CommentCount)
		}
	})

	t.Run("Tier with user relation", func(t *testing.T) {
		tier := Tier{
			UserID:   user.ID,
			Platform: "Koyeb",
			Name:     "Koyeb Free",
		}
		db.Create(&tier)

		var fetchedTier Tier
		db.Preload("User").First(&fetchedTier, tier.ID)

		if fetchedTier.User.ID != user.ID {
			t.Error("User relation not loaded correctly")
		}
		if fetchedTier.User.Username != "tieruser" {
			t.Errorf("Expected username 'tieruser', got '%s'", fetchedTier.User.Username)
		}
	})
}

func TestVoteModel(t *testing.T) {
	db := setupTestDB(t)

	// Setup test data
	user := User{Username: "voter", Email: "voter@example.com"}
	db.Create(&user)

	tier := Tier{UserID: user.ID, Platform: "Railway", Name: "Test Tier"}
	db.Create(&tier)

	t.Run("Create vote", func(t *testing.T) {
		vote := Vote{
			UserID:   user.ID,
			TierID:   tier.ID,
			VoteType: 1, // upvote
		}

		result := db.Create(&vote)
		if result.Error != nil {
			t.Errorf("Failed to create vote: %v", result.Error)
		}
	})

	t.Run("Unique vote constraint", func(t *testing.T) {
		vote1 := Vote{UserID: user.ID, TierID: tier.ID, VoteType: 1}
		db.Create(&vote1)

		vote2 := Vote{UserID: user.ID, TierID: tier.ID, VoteType: -1}
		result := db.Create(&vote2)

		if result.Error == nil {
			t.Error("Expected error for duplicate vote from same user on same tier")
		}
	})

	t.Run("Valid vote types", func(t *testing.T) {
		// Create another tier for this test
		tier2 := Tier{UserID: user.ID, Platform: "Vercel", Name: "Test Tier 2"}
		db.Create(&tier2)

		upvote := Vote{UserID: user.ID, TierID: tier2.ID, VoteType: 1}
		if err := db.Create(&upvote).Error; err != nil {
			t.Errorf("Upvote should be valid: %v", err)
		}

		// Create another user for downvote test
		user2 := User{Username: "voter2", Email: "voter2@example.com"}
		db.Create(&user2)

		downvote := Vote{UserID: user2.ID, TierID: tier2.ID, VoteType: -1}
		if err := db.Create(&downvote).Error; err != nil {
			t.Errorf("Downvote should be valid: %v", err)
		}
	})
}

func TestCommentModel(t *testing.T) {
	db := setupTestDB(t)

	// Setup test data
	user := User{Username: "commenter", Email: "commenter@example.com"}
	db.Create(&user)

	tier := Tier{UserID: user.ID, Platform: "Railway", Name: "Test Tier"}
	db.Create(&tier)

	t.Run("Create comment", func(t *testing.T) {
		comment := Comment{
			UserID:  user.ID,
			TierID:  tier.ID,
			Content: "This is a great tier!",
		}

		result := db.Create(&comment)
		if result.Error != nil {
			t.Errorf("Failed to create comment: %v", result.Error)
		}
	})

	t.Run("Comment with max length", func(t *testing.T) {
		longContent := "This comment is exactly one hundred characters long including all spaces, punctuation and more text!"
		if len(longContent) != 100 {
			t.Fatalf("Test content should be exactly 100 chars, got %d", len(longContent))
		}

		comment := Comment{
			UserID:  user.ID,
			TierID:  tier.ID,
			Content: longContent,
		}

		result := db.Create(&comment)
		if result.Error != nil {
			t.Errorf("Failed to create comment with 100 chars: %v", result.Error)
		}
	})

	t.Run("Comment timestamps", func(t *testing.T) {
		beforeCreate := time.Now()
		comment := Comment{
			UserID:  user.ID,
			TierID:  tier.ID,
			Content: "Testing timestamps",
		}
		db.Create(&comment)
		afterCreate := time.Now()

		if comment.CreatedAt.Before(beforeCreate) || comment.CreatedAt.After(afterCreate) {
			t.Error("CreatedAt timestamp is not set correctly")
		}
	})

	t.Run("Comment with relations", func(t *testing.T) {
		comment := Comment{
			UserID:  user.ID,
			TierID:  tier.ID,
			Content: "Testing relations",
		}
		db.Create(&comment)

		var fetchedComment Comment
		db.Preload("User").Preload("Tier").First(&fetchedComment, comment.ID)

		if fetchedComment.User.Username != "commenter" {
			t.Error("User relation not loaded correctly")
		}
		if fetchedComment.Tier.Name != "Test Tier" {
			t.Error("Tier relation not loaded correctly")
		}
	})
}

func TestSoftDelete(t *testing.T) {
	db := setupTestDB(t)

	t.Run("Soft delete user", func(t *testing.T) {
		user := User{Username: "deleteme", Email: "delete@example.com"}
		db.Create(&user)

		db.Delete(&user)

		var count int64
		db.Model(&User{}).Where("id = ?", user.ID).Count(&count)
		if count != 0 {
			t.Error("Deleted user should not be found in normal queries")
		}

		// Check with Unscoped
		db.Unscoped().Model(&User{}).Where("id = ?", user.ID).Count(&count)
		if count != 1 {
			t.Error("Deleted user should be found with Unscoped")
		}
	})
}
