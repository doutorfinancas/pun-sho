package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Shorty struct {
	ID             uuid.UUID       `json:"id" gorm:"column:id"`
	PublicID       string          `json:"-" gorm:"column:public_id"`
	Link           string          `json:"link" gorm:"column:link"`
	TTL            *time.Time      `json:"TTL" gorm:"column:ttl"`
	CreatedAt      *time.Time      `json:"created_at" gorm:"column:created_at"`
	DeletedAt      *time.Time      `json:"deleted_at" gorm:"column:deleted_at"`
	ShortyAccesses []*ShortyAccess `json:"accesses" gorm:"-"`
	ShortLink      string          `json:"short_link" gorm:"-"`
}

func (*Shorty) TableName() string {
	return "shorties"
}

func (s *Shorty) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New()

	return
}
