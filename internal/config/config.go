package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	TMDBToken      string
	Env            string
	APIKey         string
	APIBaseURL     string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	tmdbToken := os.Getenv("TMDB_READ_ACCESS_TOKEN")
	if tmdbToken == "" {
		log.Fatal("TMDB_READ_ACCESS_TOKEN environment variable is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	apiKey := os.Getenv("DAOBAN_API_KEY")
	
	apiBaseURL := os.Getenv("API_BASE_URL")
	if apiBaseURL == "" {
		apiBaseURL = "http://localhost:" + port
	}

	return &Config{
		Port:           port,
		TMDBToken:      tmdbToken,
		Env:            env,
		APIKey:         apiKey,
		APIBaseURL:     apiBaseURL,
	}
}
