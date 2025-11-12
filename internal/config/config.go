package config

import (
	"os"
)

type Config struct {
	ServerPort string
	DBUrl      string
}

func Load() *Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	db := os.Getenv("DATABASE_URL")
	// if db == "" {
	// 	log.Fatal("DATABASE_URL not set")
	// }

	return &Config{
		ServerPort: port,
		DBUrl:      db,
	}
}
