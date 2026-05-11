package config

import "os"

type Config struct {
	DatabaseURL    string
	JWTSecret      string
	ServerPort     string
	AllowedOrigins string
	BaseURL        string
}

func Load() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://errormonitor:changeme@localhost:5432/errormonitor?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "change-this-in-production-min-32-chars"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000,http://localhost"),
		BaseURL:        getEnv("BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
