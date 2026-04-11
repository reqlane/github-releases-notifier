package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBUser        string
	DBPassword    string
	DBName        string
	DBHost        string
	DBPort        string
	ServerPort    string
	ServerBaseURL string
	GithubToken   string
	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string
}

func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func Load() (*Config, error) {
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return nil, errors.New("invalid smtp port environment variable")
	}

	cfg := &Config{
		DBUser:        os.Getenv("DB_USER"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        os.Getenv("DB_NAME"),
		DBHost:        os.Getenv("DB_HOST"),
		DBPort:        os.Getenv("DB_PORT"),
		ServerPort:    os.Getenv("SERVER_PORT"),
		ServerBaseURL: os.Getenv("SERVER_BASE_URL"),
		GithubToken:   os.Getenv("GITHUB_API_TOKEN"),
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      port,
		SMTPUsername:  os.Getenv("SMTP_USERNAME"),
		SMTPPassword:  os.Getenv("SMTP_PASSWORD"),
	}

	if cfg.DBUser == "" || cfg.DBPassword == "" ||
		cfg.DBName == "" || cfg.DBHost == "" ||
		cfg.DBPort == "" || cfg.ServerPort == "" ||
		cfg.ServerBaseURL == "" || cfg.SMTPHost == "" ||
		cfg.SMTPUsername == "" || cfg.SMTPPassword == "" {
		return nil, errors.New("missing required environment variables")
	}

	return cfg, nil
}
