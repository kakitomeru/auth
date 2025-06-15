package repository

import (
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	User    UserRepository
	Session SessionRepository
}

func NewRepository(db *gorm.DB, sessionExp time.Duration) *Repository {
	return &Repository{
		User:    NewUserRepository(db),
		Session: NewSessionRepository(db, sessionExp),
	}
}
