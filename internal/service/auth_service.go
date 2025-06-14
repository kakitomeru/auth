package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kakitomeru/auth/internal/model"
	"github.com/kakitomeru/auth/internal/repository"
	"github.com/kakitomeru/auth/pkg/bcrypt"
	"github.com/kakitomeru/auth/pkg/crypto"
	"github.com/kakitomeru/shared/config"
	"github.com/kakitomeru/shared/jwt"
	"github.com/kakitomeru/shared/telemetry"
)

var (
	ErrInvalidPassword   = errors.New("password is invalid")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrSessionExpired    = errors.New("session expired")
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (*uuid.UUID, error)
	Login(ctx context.Context, email, password string) (*TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error)
}

type AuthServiceImpl struct {
	repo *repository.Repository
	cfg  *config.Jwt
}

func NewAuthService(repo *repository.Repository, cfg *config.Jwt) *AuthServiceImpl {
	return &AuthServiceImpl{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *AuthServiceImpl) Register(
	ctx context.Context,
	username, email, password string,
) (*uuid.UUID, error) {
	// TODO: ensure email is not already in use

	ctx, span := telemetry.StartSpan(ctx, "AuthService.Register")
	defer span.End()

	user := &model.User{
		Username: username,
		Email:    email,
	}

	passwordHash, err := bcrypt.Hash(ctx, password)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, ErrInvalidPassword
	}
	user.Password = passwordHash

	userID, err := s.repo.User.Create(ctx, user)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, err
	}

	return userID, nil
}

func (s *AuthServiceImpl) Login(
	ctx context.Context,
	email, password string,
) (*TokenPair, error) {
	ctx, span := telemetry.StartSpan(ctx, "AuthService.Login")
	defer span.End()

	user, err := s.repo.User.GetByEmail(ctx, email)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, err
	}

	ok := bcrypt.Compare(ctx, password, user.Password)
	if !ok {
		telemetry.RecordError(span, ErrIncorrectPassword)
		return nil, ErrIncorrectPassword
	}

	accessToken, err := jwt.CreateJWT(user.ID, s.cfg.Exp)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, err
	}
	refreshToken := crypto.GenerateRandomBase64(64)

	session := &model.Session{
		RefreshToken: refreshToken,
		UserID:       user.ID,
	}

	if err := s.repo.Session.Create(ctx, session); err != nil {
		telemetry.RecordError(span, err)
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServiceImpl) RefreshToken(
	ctx context.Context,
	refreshToken string,
) (string, error) {
	ctx, span := telemetry.StartSpan(ctx, "AuthService.RefreshToken")
	defer span.End()

	session, err := s.repo.Session.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		telemetry.RecordError(span, err)
		return "", err
	}

	if session.ExpiresAt.Before(time.Now()) {
		err := s.repo.Session.DeleteExpiredByUserID(ctx, session.UserID)
		if err != nil {
			telemetry.RecordError(span, err)
			return "", err
		}

		return "", ErrSessionExpired
	}

	accessToken, err := jwt.CreateJWT(session.UserID, s.cfg.Exp)
	if err != nil {
		telemetry.RecordError(span, err)
		return "", err
	}

	return accessToken, nil
}

func (s *AuthServiceImpl) Logout(
	ctx context.Context,
	userID uuid.UUID,
) error {
	ctx, span := telemetry.StartSpan(ctx, "AuthService.Logout")
	defer span.End()

	return s.repo.Session.DeleteByUserID(ctx, userID)
}

func (s *AuthServiceImpl) GetUser(
	ctx context.Context,
	userID uuid.UUID,
) (*model.User, error) {
	ctx, span := telemetry.StartSpan(ctx, "AuthService.GetUser")
	defer span.End()

	return s.repo.User.GetById(ctx, userID)
}
