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

func Connect(c *Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	switch c.DatabaseType {
	case PostGreType:
		db, err = gorm.Open(postgres.Open(*c.ConnectionString()), &gorm.Config{})
	case MySQLType:
		db, err = gorm.Open(mysql.Open(*c.ConnectionString()), &gorm.Config{})
	}

	dbConfig, _ := db.DB()
	dbConfig.SetMaxIdleConns(MaxIdle)
	dbConfig.SetMaxOpenConns(MaxConnections)
	dbConfig.SetConnMaxLifetime(time.Hour)

	return db, err
}
