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

func NewConfig(envpaths ...string) *Config {
	err := dotenv.Load(envpaths...)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		GitHubAccessToken: os.Getenv("GITHUB_ACCESS_TOKEN"),
		RepoAPI:           os.Getenv("REPO_API"),
		GASAPI:            os.Getenv("GAS_API"),
		SMTPFrom:          os.Getenv("SMTP_FROM"),
		SMTPPass:          os.Getenv("SMTP_PASS"),
		SMTPHost:          os.Getenv("SMTP_HOST"),
		SMTPPort:          os.Getenv("SMTP_PORT"),
		GEMINI_API_KEY:    os.Getenv("GEMINI_API_KEY"),
		ADMIN_MAIL:        os.Getenv("ADMIN_MAIL"),
	}
}
