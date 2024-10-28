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
	_ = dotenv.Load(envpaths...)

	return &Config{
		GitHubAccessToken: readEnvOrPanic("GITHUB_ACCESS_TOKEN"),
		RepoAPI:           readEnvOrPanic("REPO_API"),
		GASAPI:            readEnvOrPanic("GAS_API"),
		SMTPFrom:          readEnvOrPanic("SMTP_FROM"),
		SMTPPass:          readEnvOrPanic("SMTP_PASS"),
		SMTPHost:          readEnvOrPanic("SMTP_HOST"),
		SMTPPort:          readEnvOrPanic("SMTP_PORT"),
		GEMINI_API_KEY:    readEnvOrPanic("GEMINI_API_KEY"),
		ADMIN_MAIL:        readEnvOrPanic("ADMIN_MAIL"),
	}

}

func readEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Panicf("Environment variable %s is not set", key)
	}
	return value
}
