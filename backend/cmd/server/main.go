package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	"github.com/NeoRecasata/film-gallery/backend/internal/api"
	"github.com/NeoRecasata/film-gallery/backend/internal/auth"
	"github.com/NeoRecasata/film-gallery/backend/internal/config"
	"github.com/NeoRecasata/film-gallery/backend/internal/db"
	"github.com/NeoRecasata/film-gallery/backend/internal/media"
	"github.com/NeoRecasata/film-gallery/backend/internal/storage"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "reset-password" {
		resetPassword()
		return
	}

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

	srv := &api.Server{
		DB:        database,
		Config:    cfg,
		Storage:   store,
		JWT:       jwtSvc,
		Processor: processor,
	}

	router := api.NewRouter(srv)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), router); err != nil {
		log.Fatal(err)
	}
}

func resetPassword() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	var email string
	for _, arg := range os.Args[2:] {
		if strings.HasPrefix(arg, "--email=") {
			email = strings.TrimPrefix(arg, "--email=")
		}
	}
	if email == "" {
		log.Fatal("Usage: server reset-password --email=user@example.com")
	}

	fmt.Print("New password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // newline after hidden input
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}
	password := string(passwordBytes)
	if password == "" {
		log.Fatal("Password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	result, err := database.Exec(
		"UPDATE users SET password_hash = $1, token_version = token_version + 1 WHERE email = $2",
		string(hash), email,
	)
	if err != nil {
		log.Fatalf("Failed to update password: %v", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		log.Fatalf("No user found with email: %s", email)
	}

	fmt.Println("Password reset successfully. All sessions have been invalidated.")
}
