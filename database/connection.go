package database

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	MaxIdle        = 0
	MaxConnections = 10
)

func Connect(c *Config, gb *gorm.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	switch c.DatabaseType {
	case PostGreType:
		db, err = gorm.Open(postgres.Open(*c.ConnectionString()), gb)
	case MySQLType:
		db, err = gorm.Open(mysql.Open(*c.ConnectionString()), gb)
	}

	dbConfig, _ := db.DB()
	dbConfig.SetMaxIdleConns(MaxIdle)
	dbConfig.SetMaxOpenConns(MaxConnections)
	dbConfig.SetConnMaxLifetime(time.Hour)

	return db, err
}
