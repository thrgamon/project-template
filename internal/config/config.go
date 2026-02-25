package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port          int
	DatabaseURL   string
	Environment   string
	SessionMaxAge time.Duration
	CookieDomain  string
	CookieSecure  bool
}

func LoadConfig() Config {
	port := 8080
	if v, ok := os.LookupEnv("PORT"); ok {
		if parsed, err := strconv.Atoi(v); err == nil {
			port = parsed
		}
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/myapp?sslmode=disable"
	}

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	sessionMaxAge := 7 * 24 * time.Hour
	if v, ok := os.LookupEnv("SESSION_MAX_AGE"); ok {
		if secs, err := strconv.Atoi(v); err == nil {
			sessionMaxAge = time.Duration(secs) * time.Second
		}
	}

	cookieSecure := environment == "production"
	if v, ok := os.LookupEnv("COOKIE_SECURE"); ok {
		cookieSecure, _ = strconv.ParseBool(v)
	}

	return Config{
		Port:          port,
		DatabaseURL:   databaseURL,
		Environment:   environment,
		SessionMaxAge: sessionMaxAge,
		CookieDomain:  os.Getenv("COOKIE_DOMAIN"),
		CookieSecure:  cookieSecure,
	}
}
