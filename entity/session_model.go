package entity

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID  `json:"id" gorm:"column:id;type:uuid;default:gen_random_uuid()"`
	UserID    uuid.UUID  `json:"user_id" gorm:"column:user_id;type:uuid"`
	Token     string     `json:"-" gorm:"column:token"`
	Verified  bool       `json:"-" gorm:"column:verified;default:false"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"column:expires_at"`
	CreatedAt *time.Time `json:"created_at" gorm:"column:created_at"`
}

func (*Session) TableName() string {
	return "sessions"
}
