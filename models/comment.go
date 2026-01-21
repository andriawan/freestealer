package models

import (
	"time"

	"gorm.io/gorm"
)

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
