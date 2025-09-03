package api

import (
	"strings"

	"github.com/doutorfinancas/pun-sho/database"
)

type Config struct {
	Port              int    `env:"API_PORT"`
	Token             string `env:"AUTH_TOKEN"`
	HostName          string `env:"HOST_NAME"`
	UnknownPage       string `env:"UNKNOWN_PAGE"`
	QRLogo            string `env:"QR_PNG_LOGO"`
	DBUsername        string `env:"DB_USERNAME"`
	DBPassword        string `env:"DB_PASSWORD"`
	DBName            string `env:"DB_NAME"`
	DBHost            string `env:"DB_URL"`
	DBPort            int    `env:"DB_PORT"`
	SSLMode           string `env:"SSL_MODE"`
	PublicIDLength    int    `env:"PUBLIC_ID_LENGTH"`
	AllowedSocialBots string `env:"ALLOWED_SOCIAL_BOTS"`
}

func (c *Config) GetDatabaseConfig() *database.Config {
	return &database.Config{
		Host:         c.DBHost,
		Port:         c.DBPort,
		Database:     c.DBName,
		User:         c.DBUsername,
		Pass:         c.DBPassword,
		SSLMode:      c.SSLMode,
		DatabaseType: database.PostGreType,
	}
}

// GetAllowedSocialBots converts the configuration string into a slice of allowed bots
func (c *Config) GetAllowedSocialBots() []string {
	if c.AllowedSocialBots == "" {
		// Default list of social media bots if not configured
		return []string{
			"facebookexternalhit",
			"facebot",
			"googlebot",
			"linkedinbot",
			"twitterbot",
			"instagram",
			"instagrambot",
			"whatsapp",
			"slackbot",
			"telegrambot",
			"discordbot",
			"pinterestbot",
			"redditbot",
			"skypeuri",
			"applebot",
			"bingbot",
			"yandexbot",
		}
	}

	bots := strings.Split(c.AllowedSocialBots, ",")
	var trimmedBots []string
	for _, bot := range bots {
		trimmed := strings.TrimSpace(bot)
		if trimmed != "" {
			trimmedBots = append(trimmedBots, trimmed)
		}
	}
	return trimmedBots
}
