package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	Port             string
	JWTSecret        string
	JWTAccessExpiry  string
	JWTRefreshExpiry string

	StarPayAPISecret      string
	StarPayCallbackSecret string
	StarPayBaseURL        string
	StarPayCallbackURL    string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	accessExpiry := os.Getenv("JWT_ACCESS_EXPIRY")
	if accessExpiry == "" {
		accessExpiry = "15m"
	}

	refreshExpiry := os.Getenv("JWT_REFRESH_EXPIRY")
	if refreshExpiry == "" {
		refreshExpiry = "168h"
	}

	starPayBaseURL := os.Getenv("STARPAY_BASE_URL")
	if starPayBaseURL == "" {
		starPayBaseURL = "https://starpayqa.starpayethiopia.com/v1/starpay-api"
	}

	return &Config{
		DatabaseURL:      databaseURL,
		Port:             port,
		JWTSecret:        jwtSecret,
		JWTAccessExpiry:  accessExpiry,
		JWTRefreshExpiry: refreshExpiry,

		StarPayAPISecret:      os.Getenv("STARPAY_API_SECRET"),
		StarPayCallbackSecret: os.Getenv("STARPAY_CALLBACK_SECRET"),
		StarPayBaseURL:        starPayBaseURL,
		StarPayCallbackURL:    os.Getenv("STARPAY_CALLBACK_URL"),
	}, nil
}
