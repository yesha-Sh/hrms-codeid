package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv            string
	HTTPPort          string
	FrontendOrigin    string
	DatabaseURL       string
	DBHost            string
	DBPort            string
	DBName            string
	DBUser            string
	DBPassword        string
	DBSSLMode         string
	JWTAccessSecret   string
	JWTRefreshSecret  string
	JWTAccessTTL      time.Duration
	JWTRefreshTTL     time.Duration
	AdminSeedEmail    string
	AdminSeedPassword string
}

func Load() (Config, error) {
	// Try loading from .env file if it exists
	_ = godotenv.Load()

	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return Config{}, fmt.Errorf("parse JWT_ACCESS_TTL: %w", err)
	}

	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse JWT_REFRESH_TTL: %w", err)
	}

	cfg := Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		HTTPPort:          getEnv("HTTP_PORT", getEnv("PORT", "8080")),
		FrontendOrigin:    getEnv("FRONTEND_ORIGIN", "http://localhost:5173"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		DBHost:            getEnv("DB_HOST", "127.0.0.1"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBName:            getEnv("DB_NAME", "hrms"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		JWTAccessSecret:   getEnv("JWT_ACCESS_SECRET", ""),
		JWTRefreshSecret:  getEnv("JWT_REFRESH_SECRET", ""),
		JWTAccessTTL:      accessTTL,
		JWTRefreshTTL:     refreshTTL,
		AdminSeedEmail:    getEnv("ADMIN_SEED_EMAIL", ""),
		AdminSeedPassword: getEnv("ADMIN_SEED_PASSWORD", ""),
	}

	if cfg.JWTAccessSecret == "" || cfg.JWTRefreshSecret == "" {
		return Config{}, fmt.Errorf("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET are required")
	}

	return cfg, nil
}

func (c Config) DSN() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=public",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
	)
}

func (c Config) Address() string {
	return ":" + c.HTTPPort
}

func (c Config) CookieSecure() bool {
	return c.AppEnv != "development"
}

func (c Config) DBPortInt() int {
	value, _ := strconv.Atoi(c.DBPort)
	return value
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
