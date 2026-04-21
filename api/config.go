package api

import (
	"strings"
	"time"

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

	// Auth & Session config
	AdminDefaultPassword   string `env:"ADMIN_DEFAULT_PASSWORD"`
	CookieDomain           string `env:"COOKIE_DOMAIN"`
	SessionDuration        string `env:"SESSION_DURATION"`

	// Login config
	DisableLocalLogin      bool   `env:"DISABLE_LOCAL_LOGIN"`

	// Microsoft OAuth config
	MicrosoftTenantID      string `env:"MICROSOFT_TENANT_ID"`
	MicrosoftClientID      string `env:"MICROSOFT_CLIENT_ID"`
	MicrosoftSecret        string `env:"MICROSOFT_SECRET"`
	MicrosoftAllowedGroups string `env:"MICROSOFT_ALLOWED_GROUPS"`

	// GeoIP config
	GeoIPDBPath            string `env:"GEOIP_DB_PATH"`
	GeoIPLicenseKey        string `env:"GEOIP_LICENSE_KEY"`
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

// GetConfiguredSocialBots converts the configuration string into a slice of configured bots
// Returns empty slice if not configured, giving full control to each environment
func (c *Config) GetConfiguredSocialBots() []string {
	if c.AllowedSocialBots != "" {
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
	return []string{} // No bots allowed by default - explicit configuration required
}

func (c *Config) GetSessionDuration() time.Duration {
	if c.SessionDuration != "" {
		d, err := time.ParseDuration(c.SessionDuration)
		if err == nil {
			return d
		}
	}
	return 48 * time.Hour
}

func (c *Config) GetMicrosoftAllowedGroups() []string {
	if c.MicrosoftAllowedGroups != "" {
		groups := strings.Split(c.MicrosoftAllowedGroups, ",")
		var result []string
		for _, g := range groups {
			trimmed := strings.TrimSpace(g)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return nil
}
