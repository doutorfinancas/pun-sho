package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id" gorm:"column:id;type:uuid;default:gen_random_uuid()"`
	Username     string     `json:"username" gorm:"column:username"`
	Email        string     `json:"email" gorm:"column:email"`
	PasswordHash string     `json:"-" gorm:"column:password_hash"`
	TOTPSecret   string     `json:"-" gorm:"column:totp_secret"`
	TOTPEnabled  bool       `json:"totp_enabled" gorm:"column:totp_enabled"`
	MSLinked     bool       `json:"ms_linked" gorm:"column:ms_linked"`
	MSEmail      string     `json:"ms_email,omitempty" gorm:"column:ms_email"`
	Role         string     `json:"role" gorm:"column:role"`
	CreatedAt    *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    *time.Time `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at"`
}

func (*User) TableName() string {
	return "users"
}
