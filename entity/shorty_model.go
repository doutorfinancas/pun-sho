package entity

import (
	"time"

	"github.com/google/uuid"
)

type Shorty struct {
	ID               uuid.UUID      `json:"id" gorm:"column:id;type:uuid;default:uuid_generate_v4()"`
	PublicID         string         `json:"-" gorm:"column:public_id"`
	Link             string         `json:"link" gorm:"column:link"`
	TTL              *time.Time     `json:"TTL" gorm:"column:ttl"`
	RedirectionLimit *int           `json:"redirection_limit" gorm:"column:redirection_limit"`
	CreatedAt        *time.Time     `json:"created_at" gorm:"column:created_at"`
	DeletedAt        *time.Time     `json:"deleted_at" gorm:"column:deleted_at"`
	ShortyAccesses   []ShortyAccess `json:"accesses" gorm:"-"`
	ShortLink        string         `json:"short_link" gorm:"-"`
	Visits           int            `json:"visits" gorm:"-"`
	RedirectCount    int            `json:"redirects" gorm:"-"`
	QRCode           string         `json:"qr_code,omitempty" gorm:"column:qr_code"`
}

func (*Shorty) TableName() string {
	return "shorties"
}

type ShortyForList struct {
	ID               uuid.UUID      `json:"id" gorm:"column:id;type:uuid;default:uuid_generate_v4()"`
	PublicID         string         `json:"public_id" gorm:"column:public_id"`
	Link             string         `json:"link" gorm:"column:link"`
	TTL              *time.Time     `json:"TTL,omitempty" gorm:"column:ttl"`
	RedirectionLimit *int           `json:"redirection_limit,omitempty" gorm:"column:redirection_limit"`
	CreatedAt        *time.Time     `json:"created_at" gorm:"column:created_at"`
	DeletedAt        *time.Time     `json:"deleted_at" gorm:"column:deleted_at"`
	ShortyAccesses   []ShortyAccess `json:"accesses,omitempty" gorm:"-"`
	ShortLink        string         `json:"short_link,omitempty" gorm:"-"`
	Visits           int            `json:"visits" gorm:"visits"`
	Redirects        int            `json:"redirects" gorm:"redirects"`
	QRCode           string         `json:"qr_code,omitempty" gorm:"column:qr_code"`
}
