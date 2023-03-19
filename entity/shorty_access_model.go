package entity

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/pgtype"
)

type ShortyAccess struct {
	BaseModel
	ShortyID        uuid.UUID    `gorm:"column:shorty_id"`
	Meta            pgtype.JSONB `gorm:"column:meta"`
	UserAgent       string       `gorm:"column:user_agent"`
	IPAddress       string       `gorm:"column:ip_address"`
	Extra           string       `gorm:"column:extra"`
	OperatingSystem string       `gorm:"column:operating_system"`
	Browser         string       `gorm:"column:browser"`
	Status          string       `gorm:"column:status"`
}

func (ShortyAccess) TableName() string {
	return "shorty_accesses"
}
