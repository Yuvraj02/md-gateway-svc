package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds API Gateway configuration. No database credentials — gateway has no DB.
type Config struct {
	AppEnv   string
	Service  string
	HTTPPort int

	AuthServiceGRPCAddr string
	BlogServiceGRPCAddr string

	CORSAllowedOrigins string
	OwnerStudioSecret  string
	ShutdownTimeout    time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		Service:            "gateway",
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
		OwnerStudioSecret:  getEnv("OWNER_STUDIO_SECRET", "marketing-digest-dev"),
	}

	var err error
	cfg.HTTPPort, err = getEnvInt("HTTP_PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	cfg.AuthServiceGRPCAddr = os.Getenv("AUTH_SERVICE_GRPC_ADDR")
	cfg.BlogServiceGRPCAddr = os.Getenv("BLOG_SERVICE_GRPC_ADDR")
	cfg.ShutdownTimeout, err = getEnvDuration("SHUTDOWN_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}

	if cfg.AuthServiceGRPCAddr == "" {
		return Config{}, fmt.Errorf("missing required env: AUTH_SERVICE_GRPC_ADDR")
	}
	if cfg.BlogServiceGRPCAddr == "" {
		return Config{}, fmt.Errorf("missing required env: BLOG_SERVICE_GRPC_ADDR")
	}
	if cfg.HTTPPort <= 0 {
		return Config{}, fmt.Errorf("HTTP_PORT must be positive")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return n, nil
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return d, nil
}
