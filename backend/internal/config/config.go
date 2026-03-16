package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string

	StorageType      string
	StorageLocalPath string

	StorageS3Bucket    string
	StorageS3Endpoint  string
	StorageS3Region    string
	StorageS3AccessKey string
	StorageS3SecretKey string
	StorageS3Public    bool

	ImageThumbWidth  int
	ImageMediumWidth int
	ImageFullWidth   int

	MigrationsPath string
}

func Load() (*Config, error) {
	godotenv.Load()

	cfg := &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		Port:               getEnvDefault("PORT", "8080"),
		StorageType:        getEnvDefault("STORAGE_TYPE", "local"),
		StorageLocalPath:   getEnvDefault("STORAGE_LOCAL_PATH", "./data/photos"),
		StorageS3Bucket:    os.Getenv("STORAGE_S3_BUCKET"),
		StorageS3Endpoint:  os.Getenv("STORAGE_S3_ENDPOINT"),
		StorageS3Region:    os.Getenv("STORAGE_S3_REGION"),
		StorageS3AccessKey: os.Getenv("STORAGE_S3_ACCESS_KEY"),
		StorageS3SecretKey: os.Getenv("STORAGE_S3_SECRET_KEY"),
		StorageS3Public:    os.Getenv("STORAGE_S3_PUBLIC") == "true",
		ImageThumbWidth:    getEnvInt("IMAGE_THUMB_WIDTH", 400),
		ImageMediumWidth:   getEnvInt("IMAGE_MEDIUM_WIDTH", 1200),
		ImageFullWidth:     getEnvInt("IMAGE_FULL_WIDTH", 2400),
		MigrationsPath:     getEnvDefault("MIGRATIONS_PATH", "./migrations"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
