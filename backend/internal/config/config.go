package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port          string
	MongoURI      string
	MongoDatabase string
	JWTSecret     string
	FrontendURL   string
}

func Load() (Config, error) {
	cfg := Config{
		Port:          os.Getenv("PORT"),
		MongoURI:      os.Getenv("MONGO_URI"),
		MongoDatabase: os.Getenv("MONGO_DATABASE"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		FrontendURL:   os.Getenv("FRONTEND_URL"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if strings.TrimSpace(cfg.FrontendURL) == "" {
		cfg.FrontendURL = "http://localhost:5173"
	}

	missing := make([]string, 0, 3)
	for name, value := range map[string]string{
		"MONGO_URI":      cfg.MongoURI,
		"MONGO_DATABASE": cfg.MongoDatabase,
		"JWT_SECRET":     cfg.JWTSecret,
	} {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	if cfg.JWTSecret == "change-me" {
		return Config{}, errors.New("JWT_SECRET must be changed from the default value")
	}

	return cfg, nil
}
