package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		return Config{}, fmt.Errorf("dont found DATABASE_URL environment")
	}

	return Config{
		Port:        port,
		DatabaseURL: dbUrl,
	}, nil
}
