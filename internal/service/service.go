package service

import (
	"time"

	"github.com/kakitomeru/auth/internal/repository"
)

type Service struct {
	Auth AuthService
}

func NewService(repo *repository.Repository, jwtExp time.Duration) *Service {
	return &Service{
		Auth: NewAuthService(repo, jwtExp),
	}
}
