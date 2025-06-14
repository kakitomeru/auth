package service

import (
	"github.com/kakitomeru/auth/internal/repository"
	"github.com/kakitomeru/shared/config"
)

type Service struct {
	Auth AuthService
}

func NewService(repo *repository.Repository, cfg *config.Jwt) *Service {
	return &Service{
		Auth: NewAuthService(repo, cfg),
	}
}
