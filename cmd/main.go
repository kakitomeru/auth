package main

import (
	"context"
	"log"

	. "github.com/kakitomeru/auth/internal/app"
	"github.com/kakitomeru/auth/internal/model"
	"github.com/kakitomeru/shared/config"
	"github.com/kakitomeru/shared/database"
	"github.com/kakitomeru/shared/env"
)

func main() {
	cfg, err := config.LoadAuth()
	if err != nil {
		log.Fatalf("failed to load auth config: %v", err)
	}

	if err := env.LoadEnv(cfg.Env); err != nil {
		log.Fatal(err.Error())
	}

	db, err := database.ConnectDatabase()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Running migration..")
	if err = database.Migrate(db, model.User{}, model.Session{}); err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Starting app...")
	app := NewApp(db, cfg)
	ctx := context.Background()

	if err := app.Start(ctx); err != nil {
		log.Fatal(err.Error())
	}
}
