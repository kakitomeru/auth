package repository

import (
	"github.com/kakitomeru/shared/config"
	"gorm.io/gorm"
)

type Repository struct {
	User    UserRepository
	Session SessionRepository
}

func NewRepository(db *gorm.DB, sessionCfg *config.Session) *Repository {
	return &Repository{
		User:    NewUserRepository(db),
		Session: NewSessionRepository(db, sessionCfg),
	}
}
