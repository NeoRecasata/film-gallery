package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/NeoRecasata/film-gallery/backend/internal/api"
	"github.com/NeoRecasata/film-gallery/backend/internal/auth"
	"github.com/NeoRecasata/film-gallery/backend/internal/config"
	"github.com/NeoRecasata/film-gallery/backend/internal/db"
	"github.com/NeoRecasata/film-gallery/backend/internal/media"
	"github.com/NeoRecasata/film-gallery/backend/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database, cfg.MigrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	store, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	jwtSvc := auth.NewJWTService(cfg.JWTSecret)

	processor := media.NewProcessor(cfg.ImageThumbWidth, cfg.ImageMediumWidth, cfg.ImageFullWidth)

	server := &api.Server{
		DB:        database,
		Config:    cfg,
		Storage:   store,
		JWT:       jwtSvc,
		Processor: processor,
	}

	router := api.NewRouter(server)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), router); err != nil {
		log.Fatal(err)
	}
}
