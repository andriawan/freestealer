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

// Vote represents an upvote or downvote on a tier
type Vote struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index:idx_user_tier,unique" json:"user_id"`
	TierID    uint           `gorm:"not null;index:idx_user_tier,unique;index" json:"tier_id"`
	VoteType  int8           `gorm:"not null" json:"vote_type"` // 1 for upvote, -1 for downvote
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Tier Tier `gorm:"foreignKey:TierID" json:"tier,omitempty"`
}

// Comment represents a comment on a tier
type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	TierID    uint           `gorm:"not null;index" json:"tier_id"`
	Content   string         `gorm:"not null;size:100" json:"content"` // max 100 characters
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Tier Tier `gorm:"foreignKey:TierID" json:"tier,omitempty"`
}
