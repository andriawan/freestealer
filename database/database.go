package database

import (
	"fmt"
	"os"

	"freestealer/models"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	defaultHost     = "localhost"
	defaultPort     = "5432"
	defaultUser     = "postgres"
	defaultPassword = "postgres"
	defaultDBName   = "freestealer"
	defaultSchema   = "public"
	defaultSSLMode  = "disable"
)

var DB *gorm.DB

// InitDatabase initializes the PostgreSQL database connection and runs migrations
func InitDatabase() error {
	var err error

	// Get database configuration from environment variables with defaults
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = defaultHost
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = defaultPort
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = defaultUser
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = defaultPassword
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = defaultDBName
	}

	schema := os.Getenv("DB_SCHEMA")
	if schema == "" {
		schema = defaultSchema
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = defaultSSLMode
	}

	// Build connection string

	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		host, port, user, password, dbname, sslmode, schema)

	// Open PostgreSQL database connection
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"host":   host,
		"port":   port,
		"dbname": dbname,
		"schema": schema,
	}).Info("Database connection established")

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
	// PostgreSQL syntax
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
