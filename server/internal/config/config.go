package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
	JWTSecret   string
	BaseURL     string
}

func Load() *Config {
	loadDotEnv(".env")

	return &Config{
		HTTPPort:    envOr("HTTP_PORT", "8080"),
		GRPCPort:    envOr("GRPC_PORT", "9090"),
		DatabaseURL: envOr("DATABASE_URL", "postgres://shortify:shortify@localhost:5432/shortify?sslmode=disable"),
		JWTSecret:   envOr("JWT_SECRET", "change-me"),
		BaseURL:     strings.TrimRight(envOr("BASE_URL", "http://localhost:8080"), "/"),
	}
}

func envOr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return
	}
}
