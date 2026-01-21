package models

import (
	"time"

	"gorm.io/gorm"
)

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
