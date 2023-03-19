package entity

import (
	"github.com/google/uuid"
)

type ShortyAccess struct {
	BaseModel
	ShortyID        uuid.UUID           `gorm:"column:shorty_id"`
	Meta            map[string][]string `gorm:"column:meta;type:JSONB"`
	UserAgent       string              `gorm:"column:user_agent"`
	IPAddress       string              `gorm:"column:ip_address"`
	Extra           string              `gorm:"column:extra"`
	OperatingSystem string              `gorm:"column:operating_system"`
	Browser         string              `gorm:"column:browser"`
	Status          string              `gorm:"column:status"`
}

func (ShortyAccess) TableName() string {
	return "shorty_accesses"
}
