package api

import (
	"github.com/doutorfinancas/pun-sho/database"
)

type Config struct {
	Port        int    `env:"API_PORT"`
	Token       string `env:"AUTH_TOKEN"`
	UnknownPage string `env:"UNKNOWN_PAGE"`
	DBUsername  string `env:"DB_USERNAME"`
	DBPassword  string `env:"DB_PASSWORD"`
	DBName      string `env:"DB_NAME"`
	DBHost      string `env:"DB_URL"`
	DBPort      int    `env:"DB_PORT"`
}

func (c *Config) GetDatabaseConfig() *database.Config {
	return &database.Config{
		Host:         c.DBHost,
		Port:         c.DBPort,
		Database:     c.DBName,
		User:         c.DBUsername,
		Pass:         c.DBPassword,
		DatabaseType: database.PostGreType,
	}
}
