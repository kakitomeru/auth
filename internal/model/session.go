package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/kakitomeru/shared/model"
)

type Session struct {
	model.Model
	RefreshToken string    `gorm:"not null;index"`
	ExpiresAt    time.Time `gorm:"not null"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`
}
