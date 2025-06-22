package main

import (
	"context"
	"log/slog"
	"os"

	. "github.com/kakitomeru/auth/internal/app"
	"github.com/kakitomeru/auth/internal/config"
	"github.com/kakitomeru/auth/pkg/model"
	"github.com/kakitomeru/shared/database"
	"github.com/kakitomeru/shared/env"
	"github.com/kakitomeru/shared/logger"
)

func main() {
	logger.InitSlog("auth", "dev", slog.LevelDebug)
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error(ctx, "failed to load auth config", err)
		os.Exit(1)
	}

	err = env.LoadEnv(cfg.Env)
	if err != nil {
		logger.Error(ctx, "failed to load env", err)
		os.Exit(1)
	}

	logger.Debug(ctx, "Connecting to database")
	db, err := database.ConnectDatabase()
	if err != nil {
		logger.Error(ctx, "failed to connect to database", err)
		os.Exit(1)
	}

	logger.Debug(ctx, "Running migration for table user and session")
	if err = database.Migrate(db, model.User{}, model.Session{}); err != nil {
		logger.Error(ctx, "failed to migrate database", err)
		os.Exit(1)
	}

	logger.Debug(ctx, "Starting app")
	app := NewApp(db, cfg)

	app.Start(ctx)
}
