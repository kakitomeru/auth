package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kakitomeru/auth/internal/model"
	"github.com/kakitomeru/shared/config"
	"github.com/kakitomeru/shared/telemetry"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type SessionRepository interface {
	Create(
		ctx context.Context,
		session *model.Session,
	) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*model.Session, error)
	DeleteExpiredByUserID(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, id uuid.UUID) error
}

type SessionRepositoryImpl struct {
	db  *gorm.DB
	cfg *config.Session
}

func NewSessionRepository(db *gorm.DB, cfg *config.Session) *SessionRepositoryImpl {
	return &SessionRepositoryImpl{
		db:  db,
		cfg: cfg,
	}
}

func (r *SessionRepositoryImpl) Create(
	ctx context.Context,
	session *model.Session,
) error {
	ctx, span := telemetry.StartSpan(ctx, "SessionRepository.Create")
	defer span.End()

	session.ID = uuid.New()
	session.ExpiresAt = time.Now().Add(r.cfg.Exp)

	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		telemetry.RecordError(span, err)
		return err
	}

	return nil
}

func (r *SessionRepositoryImpl) GetByRefreshToken(
	ctx context.Context,
	refreshToken string,
) (*model.Session, error) {
	ctx, span := telemetry.StartSpan(ctx, "SessionRepository.GetByRefreshToken")
	defer span.End()

	session := new(model.Session)

	if err := r.db.WithContext(ctx).Where("refresh_token = ?", refreshToken).First(session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			telemetry.RecordError(span, ErrSessionNotFound)
			return nil, ErrSessionNotFound
		}

		telemetry.RecordError(span, err)
		return nil, err
	}

	return session, nil
}

func (r *SessionRepositoryImpl) DeleteExpiredByUserID(
	ctx context.Context,
	id uuid.UUID,
) error {
	ctx, span := telemetry.StartSpan(ctx, "SessionRepository.DeleteExpiredByUserID")
	defer span.End()

	err := r.db.
		WithContext(ctx).
		Where("user_id = ? and expires_at < ?", id, time.Now()).
		Delete(&model.Session{}).
		Error

	if err != nil {
		telemetry.RecordError(span, err)
		return err
	}

	return nil
}

func (r *SessionRepositoryImpl) DeleteByUserID(
	ctx context.Context,
	id uuid.UUID,
) error {
	ctx, span := telemetry.StartSpan(ctx, "SessionRepository.DeleteByUserID")
	defer span.End()

	err := r.db.WithContext(ctx).Where("user_id = ?", id).Delete(&model.Session{}).Error
	if err != nil {
		telemetry.RecordError(span, err)
		return err
	}

	return nil
}
