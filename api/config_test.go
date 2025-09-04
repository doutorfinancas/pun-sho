package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_GetConfiguredSocialBots(t *testing.T) {
	tests := []struct {
		name              string
		allowedSocialBots string
		expected          []string
	}{
		{
			name:              "Empty configuration should return empty slice",
			allowedSocialBots: "",
			expected:          []string{},
		},
		{
			name:              "Single bot configuration",
			allowedSocialBots: "facebookexternalhit",
			expected:          []string{"facebookexternalhit"},
		},
		{
			name:              "Multiple bots configuration",
			allowedSocialBots: "facebookexternalhit,googlebot,linkedinbot",
			expected:          []string{"facebookexternalhit", "googlebot", "linkedinbot"},
		},
		{
			name:              "Configuration with spaces should be trimmed",
			allowedSocialBots: " facebookexternalhit , googlebot , linkedinbot ",
			expected:          []string{"facebookexternalhit", "googlebot", "linkedinbot"},
		},
		{
			name:              "Configuration with empty values should be filtered",
			allowedSocialBots: "facebookexternalhit,,googlebot,",
			expected:          []string{"facebookexternalhit", "googlebot"},
		},
		{
			name:              "Instagram bots configuration",
			allowedSocialBots: "instagram,instagrambot",
			expected:          []string{"instagram", "instagrambot"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				AllowedSocialBots: tt.allowedSocialBots,
			}
			result := config.GetConfiguredSocialBots()
			assert.Equal(t, tt.expected, result)
		})
	}
}
