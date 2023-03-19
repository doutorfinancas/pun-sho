package database

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/doutorfinancas/pun-sho/str"
)

func TestConfig_ConnectionString(t *testing.T) {
	type fields struct {
		Host     string
		Port     int
		Database string
		User     string
		Pass     string
	}
	tests := []struct {
		name   string
		fields fields
		want   *string
	}{
		{
			"Connection Successful",
			fields{
				"192.168.0.1",
				3306,
				"test",
				"root",
				"test123",
			},
			str.ToStringNil("postgresql://root:test123@192.168.0.1:3306/test?sslmode=verify-full"),
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
				}

				assert.Equal(t, tt.want, c.ConnectionString())
			},
		)
	}
}
