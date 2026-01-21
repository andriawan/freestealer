package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email    string `gorm:"uniqueIndex;not null;size:100" json:"email"`
	Password string `gorm:"size:255" json:"-"` // Hashed password, hidden from JSON

	// GitHub OAuth fields
	GitHubID     string `gorm:"size:50" json:"github_id,omitempty"` // Unique index created manually in database.go
	GitHubLogin  string `gorm:"size:100" json:"github_login,omitempty"`
	AvatarURL    string `gorm:"size:500" json:"avatar_url,omitempty"`
	AccessToken  string `gorm:"size:500" json:"-"` // Hidden from JSON
	RefreshToken string `gorm:"size:500" json:"-"` // Hidden from JSON

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Tiers    []Tier    `gorm:"foreignKey:UserID" json:"tiers,omitempty"`
	Votes    []Vote    `gorm:"foreignKey:UserID" json:"votes,omitempty"`
	Comments []Comment `gorm:"foreignKey:UserID" json:"comments,omitempty"`
}
