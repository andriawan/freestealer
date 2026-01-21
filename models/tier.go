package models

import (
	"time"

	"gorm.io/gorm"
)

// Tier represents a free tier hosting platform information
type Tier struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	UserID      uint   `gorm:"not null;index" json:"user_id"`
	Platform    string `gorm:"not null;size:100;index" json:"platform"` // e.g., Railway, Koyeb, Vercel
	Name        string `gorm:"not null;size:200" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	IsPublic    bool   `gorm:"default:true;index" json:"is_public"`

	// Tier details
	CPULimit       string `gorm:"size:50" json:"cpu_limit"`
	MemoryLimit    string `gorm:"size:50" json:"memory_limit"`
	StorageLimit   string `gorm:"size:50" json:"storage_limit"`
	BandwidthLimit string `gorm:"size:50" json:"bandwidth_limit"`
	MonthlyHours   string `gorm:"size:50" json:"monthly_hours"`
	URL            string `gorm:"size:500" json:"url"`

	// Stats (denormalized for performance)
	UpvoteCount   int `gorm:"default:0;index" json:"upvote_count"`
	DownvoteCount int `gorm:"default:0" json:"downvote_count"`
	CommentCount  int `gorm:"default:0" json:"comment_count"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Votes    []Vote    `gorm:"foreignKey:TierID" json:"votes,omitempty"`
	Comments []Comment `gorm:"foreignKey:TierID" json:"comments,omitempty"`
}
