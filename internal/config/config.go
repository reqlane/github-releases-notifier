package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	DBUser      string
	DBPassword  string
	DBName      string
	DBHost      string
	DBPort      string
	ServerPort  string
	GithubToken string
}

func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func Load() (*Config, error) {
	cfg := &Config{
		DBUser:      os.Getenv("DB_USER"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
		DBHost:      os.Getenv("DB_HOST"),
		DBPort:      os.Getenv("DB_PORT"),
		ServerPort:  os.Getenv("SERVER_PORT"),
		GithubToken: os.Getenv("GITHUB_API_TOKEN"),
	}

	if cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBName == "" || cfg.DBHost == "" || cfg.DBPort == "" || cfg.ServerPort == "" {
		return nil, errors.New("missing required environment variables")
	}

	return cfg, nil
}
