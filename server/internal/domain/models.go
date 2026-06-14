package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Link struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	OriginalURL string    `gorm:"not null" json:"original_url"`
	ShortCode   string    `gorm:"uniqueIndex;size:16;not null" json:"short_code"`
	CreatedAt   time.Time `json:"created_at"`
}

type Click struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	LinkID    uuid.UUID `gorm:"type:uuid;index;not null" json:"link_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}
