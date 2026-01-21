package database

import (
	"freestealer/models"

	"github.com/glebarez/sqlite"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the SQLite database connection and runs migrations
func InitDatabase(dbPath string) error {
	var err error

	// Open SQLite database with pure Go driver (glebarez/sqlite - no CGO required)
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return err
	}

	log.Info("Database connection established")

	// Run auto migrations
	err = DB.AutoMigrate(
		&models.User{},
		&models.Tier{},
		&models.Vote{},
		&models.Comment{},
	)

	if err != nil {
		return err
	}

	log.Info("Database migrations completed")

	// Create indexes for better performance
	createIndexes()

	return nil
}

// createIndexes creates additional composite indexes for query optimization
func createIndexes() {
	// Partial unique index for GitHubID (only when not empty)
	DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_users_git_hub_id ON users(git_hub_id) WHERE git_hub_id != ''")

	// Index for querying public tiers sorted by votes
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_tiers_public_votes ON tiers(is_public, upvote_count DESC) WHERE deleted_at IS NULL")

	// Index for querying tiers by platform
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_tiers_platform_public ON tiers(platform, is_public) WHERE deleted_at IS NULL")

	log.Info("Custom indexes created")
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
