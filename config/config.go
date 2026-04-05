package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	BaseURL string

	GitHubClientID     string
	GitHubClientSecret string

	JiraClientID     string
	JiraClientSecret string

	DatabaseURL string
}

var AppConfig *Config

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system env")
	}

	AppConfig = &Config{
		Port: getEnv("PORT", "8080"),
		BaseURL: getEnv("BASE_URL", "http://localhost:8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}