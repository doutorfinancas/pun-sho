package database

import (
	"fmt"

	"github.com/doutorfinancas/pun-sho/str"
)

const (
	postGreConnection = "postgresql://%s:%s@%s:%d/%s?sslmode=verify-full"
	mySQLConnection   = "%s:%s@tcp(%s:%d)/%s?query"
)

const (
	PostGreType = iota
	MySQLType
)

type Config struct {
	Host         string
	Port         int
	Database     string
	User         string
	Pass         string
	DatabaseType int
}

func (c *Config) ConnectionString() *string {
	var connString string
	switch c.DatabaseType {
	case PostGreType:
		connString = postGreConnection
	case MySQLType:
		connString = mySQLConnection
	}
	return str.ToStringNil(
		fmt.Sprintf(
			connString,
			c.User,
			c.Pass,
			c.Host,
			c.Port,
			c.Database,
		),
	)
}
