package service

import (
	"github.com/kakitomeru/shared/config"
	"nota.auth/internal/repository"
)

type Service struct {
	Auth AuthService
}

func NewService(repo *repository.Repository, cfg *config.Jwt) *Service {
	return &Service{
		Auth: NewAuthService(repo, cfg),
	}
}
