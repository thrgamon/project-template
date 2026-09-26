package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                 int
	DatabaseURL          string
	Environment          string
	StaticDir            string
	SessionMaxAge        time.Duration
	CookieDomain         string
	CookieSecure         bool
	AllowInsecureCookies bool
	Auth0IssuerURL       string
	Auth0ClientID        string
	Auth0ClientSecret    string
	Auth0RedirectURL     string
	AuthStateSecret      string
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
	allowInsecureCookies, _ := strconv.ParseBool(os.Getenv("AUTH_ALLOW_INSECURE_COOKIES"))

	return Config{
		Port:                 port,
		DatabaseURL:          databaseURL,
		Environment:          environment,
		StaticDir:            os.Getenv("STATIC_DIR"),
		SessionMaxAge:        sessionMaxAge,
		CookieDomain:         os.Getenv("COOKIE_DOMAIN"),
		CookieSecure:         cookieSecure,
		AllowInsecureCookies: allowInsecureCookies,
		Auth0IssuerURL:       os.Getenv("AUTH0_ISSUER_URL"),
		Auth0ClientID:        os.Getenv("AUTH0_CLIENT_ID"),
		Auth0ClientSecret:    os.Getenv("AUTH0_CLIENT_SECRET"),
		Auth0RedirectURL:     os.Getenv("AUTH0_REDIRECT_URL"),
		AuthStateSecret:      os.Getenv("AUTH_STATE_SECRET"),
	}
}
