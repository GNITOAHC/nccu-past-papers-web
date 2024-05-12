package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GitHubAccessToken string
	RepoAPI           string
	GASAPI            string
}

func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		GitHubAccessToken: os.Getenv("GITHUB_ACCESS_TOKEN"),
		RepoAPI:           os.Getenv("REPO_API"),
		GASAPI:            os.Getenv("GAS_API"),
	}
}
