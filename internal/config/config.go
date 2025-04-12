package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL         string
	JWTSecret           string
	ServerPort          string
	AccessTokenDuration time.Duration
}

func LoadConfig() (*Config, error) {

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	accessTokenDurationStr := os.Getenv("ACCESS_TOKEN_DURATION")
	if accessTokenDurationStr == "" {
		log.Fatal("ACCESS_TOKEN_DURATION environment variable is required")
	}

	accessTokenDuration, err := time.ParseDuration(accessTokenDurationStr)
	if err != nil {
		log.Fatalf("Invalid ACCESS_TOKEN_DURATION format: %v", err)
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	if _, err := strconv.Atoi(serverPort); err != nil {
		log.Fatalf("Invalid SERVER_PORT: %v", err)
	}

	return &Config{
		DatabaseURL:         dbURL,
		JWTSecret:           jwtSecret,
		ServerPort:          serverPort,
		AccessTokenDuration: accessTokenDuration,
	}, nil
}
