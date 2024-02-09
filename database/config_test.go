package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ConnectionString(t *testing.T) {
	type fields struct {
		Host     string
		Port     int
		Database string
		User     string
		Pass     string
		Mode     string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"Connection Successful",
			fields{
				"192.168.0.1",
				3306,
				"test",
				"root",
				"test123",
				"disable",
			},
			"postgresql://root:test123@192.168.0.1:3306/test?sslmode=disable",
		},
		{
			"Connection Successful with default, gets full-verify",
			fields{
				"192.168.0.1",
				3306,
				"test",
				"root",
				"test123",
				"",
			},
			"postgresql://root:test123@192.168.0.1:3306/test?sslmode=full-verify",
		},
		{
			"Connection Successful with full verify",
			fields{
				"192.168.0.1",
				3306,
				"test",
				"root",
				"test123",
				"full-verify",
			},
			"postgresql://root:test123@192.168.0.1:3306/test?sslmode=full-verify",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := &Config{
					Host:     tt.fields.Host,
					Port:     tt.fields.Port,
					Database: tt.fields.Database,
					User:     tt.fields.User,
					Pass:     tt.fields.Pass,
					SSLMode:  tt.fields.Mode,
				}

				assert.Equal(t, tt.want, *c.ConnectionString())
			},
		)
	}
}
