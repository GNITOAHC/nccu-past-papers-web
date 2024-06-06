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
}

func NewConfig() *Config {
	err := dotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		GitHubAccessToken: os.Getenv("GITHUB_ACCESS_TOKEN"),
		RepoAPI:           os.Getenv("REPO_API"),
		GASAPI:            os.Getenv("GAS_API"),
	}
}
