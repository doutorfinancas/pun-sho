package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID  ` gorm:"column:id"`
	CreatedAt *time.Time `gorm:"column:created_at"`
}

func (b *BaseModel) BeforeCreate(_ *gorm.DB) (err error) {
	b.ID = uuid.New()

	return
}
