package config

import (
	"log"
	"os"

	"past-papers-web/dotenv"
)

type Config struct {
	GitHubAccessToken string
	RepoAPI           string
	GASAPI            string
	SMTPFrom          string
	SMTPPass          string
	SMTPHost          string
	SMTPPort          string
	GEMINI_API_KEY    string
	ADMIN_MAIL        string
}

func must(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment variable %s is required", key)
	}
	return val
}

func NewConfig(envpaths ...string) *Config {
	err := dotenv.Load(envpaths...)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		GitHubAccessToken: must("GITHUB_ACCESS_TOKEN"),
		RepoAPI:           must("REPO_API"),
		GASAPI:            must("GAS_API"),
		SMTPFrom:          must("SMTP_FROM"),
		SMTPPass:          must("SMTP_PASS"),
		SMTPHost:          must("SMTP_HOST"),
		SMTPPort:          must("SMTP_PORT"),
		GEMINI_API_KEY:    must("GEMINI_API_KEY"),
		ADMIN_MAIL:        must("ADMIN_MAIL"),
	}
}
