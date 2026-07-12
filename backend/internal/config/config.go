package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName    string
	AppVersion string
	ServerPort string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	JWTSecret         string
	JWTIssuer         string
	JWTAudience       string
	JWTAccessTokenTTL time.Duration

	AuthCookieName     string
	AuthCookieSecure   bool
	AuthCookieSameSite string
	AuthCookieDomain   string

	CORSAllowedOrigins []string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado. Utilizando variáveis do ambiente.")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if strings.TrimSpace(jwtSecret) == "" {
		log.Fatal("JWT_SECRET é obrigatório e não pode estar vazio")
	}

	cfg := &Config{
		AppName:    os.Getenv("APP_NAME"),
		AppVersion: os.Getenv("APP_VERSION"),
		ServerPort: os.Getenv("SERVER_PORT"),

		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),

		JWTSecret:         jwtSecret,
		JWTIssuer:         os.Getenv("JWT_ISSUER"),
		JWTAudience:       os.Getenv("JWT_AUDIENCE"),
		JWTAccessTokenTTL: parseTTLMinutes(os.Getenv("JWT_ACCESS_TOKEN_TTL_MINUTES"), 15),

		AuthCookieName:     envOrDefault("AUTH_COOKIE_NAME", "access_token"),
		AuthCookieSecure:   parseBool(os.Getenv("AUTH_COOKIE_SECURE"), false),
		AuthCookieSameSite: envOrDefault("AUTH_COOKIE_SAME_SITE", "lax"),
		AuthCookieDomain:   os.Getenv("AUTH_COOKIE_DOMAIN"),

		CORSAllowedOrigins: parseOrigins(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}

	return cfg
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

func parseBool(value string, fallback bool) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func parseTTLMinutes(value string, fallbackMinutes int) time.Duration {
	minutes, err := strconv.Atoi(value)
	if err != nil || minutes <= 0 {
		minutes = fallbackMinutes
	}

	return time.Duration(minutes) * time.Minute
}

func parseOrigins(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}

	rawOrigins := strings.Split(value, ",")
	origins := make([]string, 0, len(rawOrigins))

	for _, origin := range rawOrigins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	return origins
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBSSLMode,
	)
}
