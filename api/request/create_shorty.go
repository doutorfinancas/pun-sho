package request

import (
	"time"
)

type CreateShorty struct {
	Link string     `json:"link" gorm:"column:link"`
	TTL  *time.Time `json:"TTL" gorm:"column:ttl"`
}
