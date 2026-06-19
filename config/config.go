package config

import (
	"os"
)

type Config struct {
	ServerPort string
	DBPath     string
}

func Load() *Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "redpacket.db"
	}
	return &Config{
		ServerPort: port,
		DBPath:     dbPath,
	}
}
