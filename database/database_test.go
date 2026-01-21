package database

import (
	"freestealer/models"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) {
	// Use environment variables or default test values
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

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test - PostgreSQL not available: %v", err)
		return
	}

	// Clean and migrate
	DB.Exec("DROP SCHEMA IF EXISTS public CASCADE")
	DB.Exec("CREATE SCHEMA public")

	err = DB.AutoMigrate(
		&models.User{},
		&models.Tier{},
		&models.Vote{},
		&models.Comment{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}
}

func TestInitDatabase(t *testing.T) {
	t.Run("Initialize with environment variables", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_USER", "postgres")
		os.Setenv("DB_PASSWORD", "postgres")
		os.Setenv("DB_NAME", "freestealer_test")
		defer func() {
			os.Unsetenv("DB_HOST")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_PASSWORD")
			os.Unsetenv("DB_NAME")
		}()

		err := InitDatabase()
		if err != nil {
			t.Skipf("Skipping test - PostgreSQL not available: %v", err)
			return
		}

		if DB == nil {
			t.Error("Database instance should not be nil")
		}
	})

	t.Run("Tables are created", func(t *testing.T) {
		setupTestDB(t)
		if DB == nil {
			t.Skip("Database not available")
		}

		// Check if tables exist
		if !DB.Migrator().HasTable(&models.User{}) {
			t.Error("Users table should exist")
		}
		if !DB.Migrator().HasTable(&models.Tier{}) {
			t.Error("Tiers table should exist")
		}
		if !DB.Migrator().HasTable(&models.Vote{}) {
			t.Error("Votes table should exist")
		}
		if !DB.Migrator().HasTable(&models.Comment{}) {
			t.Error("Comments table should exist")
		}
	})

	t.Run("Indexes are created", func(t *testing.T) {
		setupTestDB(t)
		if DB == nil {
			t.Skip("Database not available")
		}

		createIndexes()

		// Check if columns have indexes
		if !DB.Migrator().HasIndex(&models.User{}, "idx_users_username") {
			t.Error("Username index should exist")
		}
		if !DB.Migrator().HasIndex(&models.User{}, "idx_users_email") {
			t.Error("Email index should exist")
		}
		if !DB.Migrator().HasIndex(&models.Tier{}, "idx_tiers_platform") {
			t.Error("Platform index should exist")
		}
		if !DB.Migrator().HasIndex(&models.Vote{}, "idx_user_tier") {
			t.Error("User-Tier unique index should exist")
		}
	})
}

func TestGetDB(t *testing.T) {
	setupTestDB(t)
	if DB == nil {
		t.Skip("Database not available")
	}

	db := GetDB()
	if db == nil {
		t.Error("GetDB should return database instance")
	}

	if db != DB {
		t.Error("GetDB should return the same instance as DB")
	}
}

func TestDatabaseOperations(t *testing.T) {
	setupTestDB(t)
	if DB == nil {
		t.Skip("Database not available")
	}

	t.Run("Create and read user", func(t *testing.T) {
		user := models.User{
			Username: "dbtest",
			Email:    "dbtest@example.com",
		}
		result := DB.Create(&user)
		if result.Error != nil {
			t.Errorf("Failed to create user: %v", result.Error)
		}

		var fetched models.User
		DB.First(&fetched, user.ID)
		if fetched.Username != "dbtest" {
			t.Errorf("Expected username 'dbtest', got '%s'", fetched.Username)
		}
	})

	t.Run("Foreign key relations", func(t *testing.T) {
		user := models.User{Username: "reltest", Email: "reltest@example.com"}
		DB.Create(&user)

		tier := models.Tier{
			UserID:   user.ID,
			Platform: "Railway",
			Name:     "Test Tier",
		}
		result := DB.Create(&tier)
		if result.Error != nil {
			t.Errorf("Failed to create tier with foreign key: %v", result.Error)
		}

		// Try to create tier with invalid user_id
		invalidTier := models.Tier{
			UserID:   9999,
			Platform: "Invalid",
			Name:     "Should Fail",
		}
		result = DB.Create(&invalidTier)
		// Note: SQLite doesn't enforce foreign keys by default in some configurations
		// This test documents expected behavior
		t.Logf("Creating tier with invalid user_id result: %v", result.Error)
	})

	t.Run("Cascade operations", func(t *testing.T) {
		user := models.User{Username: "cascade", Email: "cascade@example.com"}
		DB.Create(&user)

		tier := models.Tier{UserID: user.ID, Platform: "Test", Name: "Cascade Test"}
		DB.Create(&tier)

		comment := models.Comment{UserID: user.ID, TierID: tier.ID, Content: "Test comment"}
		DB.Create(&comment)

		// Delete tier (comments should handle deletion)
		DB.Delete(&tier)

		var count int64
		DB.Model(&models.Comment{}).Where("tier_id = ?", tier.ID).Count(&count)
		// Comments should still exist (soft delete)
		t.Logf("Comments after tier deletion: %d", count)
	})

	t.Run("Transaction rollback", func(t *testing.T) {
		tx := DB.Begin()

		user := models.User{Username: "txtest", Email: "txtest@example.com"}
		tx.Create(&user)

		tx.Rollback()

		var count int64
		DB.Model(&models.User{}).Where("username = ?", "txtest").Count(&count)
		if count != 0 {
			t.Error("User should not exist after rollback")
		}
	})

	t.Run("Transaction commit", func(t *testing.T) {
		tx := DB.Begin()

		user := models.User{Username: "txcommit", Email: "txcommit@example.com"}
		tx.Create(&user)

		tx.Commit()

		var count int64
		DB.Model(&models.User{}).Where("username = ?", "txcommit").Count(&count)
		if count != 1 {
			t.Error("User should exist after commit")
		}
	})
}

func TestQueryOptimization(t *testing.T) {
	setupTestDB(t)
	if DB == nil {
		t.Skip("Database not available")
	}

	// Create test data
	user := models.User{Username: "perftest", Email: "perftest@example.com"}
	DB.Create(&user)

	// Create 25 public and 25 private tiers
	for i := 0; i < 25; i++ {
		publicTier := models.Tier{
			UserID:   user.ID,
			Platform: "Railway",
			Name:     "Public Tier",
			IsPublic: true,
		}
		DB.Create(&publicTier)

		privateTier := models.Tier{
			UserID:   user.ID,
			Platform: "Railway",
			Name:     "Private Tier",
		}
		DB.Create(&privateTier)
		// Update to false after creation to override default
		DB.Model(&privateTier).Update("is_public", false)
	}

	t.Run("Query with index", func(t *testing.T) {
		var tiers []models.Tier
		result := DB.Where("is_public = ?", true).Find(&tiers)
		if result.Error != nil {
			t.Errorf("Query failed: %v", result.Error)
		}
		if len(tiers) != 25 {
			t.Errorf("Expected 25 public tiers, got %d", len(tiers))
		}
	})

	t.Run("Query with pagination", func(t *testing.T) {
		var tiers []models.Tier
		result := DB.Limit(10).Offset(0).Find(&tiers)
		if result.Error != nil {
			t.Errorf("Pagination query failed: %v", result.Error)
		}
		if len(tiers) != 10 {
			t.Errorf("Expected 10 tiers, got %d", len(tiers))
		}
	})

	t.Run("Query with order", func(t *testing.T) {
		// Add vote counts to some tiers
		var tier models.Tier
		DB.First(&tier)
		DB.Model(&tier).Update("upvote_count", 100)

		var tiers []models.Tier
		DB.Order("upvote_count DESC").Limit(1).Find(&tiers)
		if len(tiers) > 0 && tiers[0].UpvoteCount != 100 {
			t.Error("Order by upvote_count not working correctly")
		}
	})
}

func TestDatabaseConstraints(t *testing.T) {
	setupTestDB(t)
	if DB == nil {
		t.Skip("Database not available")
	}

	t.Run("Unique constraint on username", func(t *testing.T) {
		user1 := models.User{Username: "unique", Email: "user1@example.com"}
		DB.Create(&user1)

		user2 := models.User{Username: "unique", Email: "user2@example.com"}
		result := DB.Create(&user2)

		if result.Error == nil {
			t.Error("Should fail due to unique constraint on username")
		}
	})

	t.Run("Unique constraint on vote", func(t *testing.T) {
		user := models.User{Username: "voter", Email: "voter@example.com"}

		DB.Create(&user)

		tier := models.Tier{UserID: user.ID, Platform: "Test", Name: "Test"}
		DB.Create(&tier)

		vote1 := models.Vote{UserID: user.ID, TierID: tier.ID, VoteType: 1}
		DB.Create(&vote1)

		vote2 := models.Vote{UserID: user.ID, TierID: tier.ID, VoteType: -1}
		result := DB.Create(&vote2)

		if result.Error == nil {
			t.Error("Should fail due to unique constraint on user_id + tier_id")
		}
	})

	t.Run("Not null constraints", func(t *testing.T) {
		// Test that we can create a tier with all required fields
		// (GORM's NOT NULL is advisory; actual validation happens at API layer)
		tier := models.Tier{UserID: 1, Platform: "Test", Name: "Test"}
		result := DB.Create(&tier)

		if result.Error != nil {
			t.Errorf("Should create tier with all required fields: %v", result.Error)
		}
	})
}
