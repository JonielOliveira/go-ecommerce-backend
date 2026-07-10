package config

import (
	"fmt"
	"log"
	"os"

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
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado. Utilizando variáveis do ambiente.")
	}

	return &Config{
		AppName:    os.Getenv("APP_NAME"),
		AppVersion: os.Getenv("APP_VERSION"),
		ServerPort: os.Getenv("SERVER_PORT"),

		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),
	}
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
