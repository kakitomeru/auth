package main

import (
	"context"
	"log"

	. "github.com/kakitomeru/auth/internal/app"
	"github.com/kakitomeru/auth/internal/config"
	"github.com/kakitomeru/auth/internal/model"
	"github.com/kakitomeru/shared/database"
	"github.com/kakitomeru/shared/env"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load auth config: %v", err)
	}

	err = env.LoadEnv(cfg.Env)
	if err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	db, err := database.ConnectDatabase()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Running migration for table user and session")
	if err = database.Migrate(db, model.User{}, model.Session{}); err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Starting app")
	app := NewApp(db, cfg)
	ctx := context.Background()

	if err := app.Start(ctx); err != nil {
		log.Fatal(err.Error())
	}
}
