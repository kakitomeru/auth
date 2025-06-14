package bcrypt

import (
	"context"

	"github.com/kakitomeru/shared/telemetry"
	"golang.org/x/crypto/bcrypt"
)

func Hash(ctx context.Context, password string) (string, error) {
	ctx, span := telemetry.StartSpan(ctx, "Bcrypt.Hash")
	defer span.End()

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		telemetry.RecordError(span, err)
		return "", err
	}

	return string(hashed), nil
}

func Compare(ctx context.Context, password, hash string) bool {
	ctx, span := telemetry.StartSpan(ctx, "Bcrypt.Compare")
	defer span.End()

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		telemetry.RecordError(span, err)
		return false
	}

	return true
}
