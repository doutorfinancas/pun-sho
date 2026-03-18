package request

import (
	"time"
)

type CreateShorty struct {
	Link             string     `json:"link" form:"link"`
	TTL              *time.Time `json:"TTL" form:"ttl"`
	RedirectionLimit *int       `json:"redirection_limit" form:"redirection_limit"`
	QRCode           *QRCode    `json:"qr_code"`
	Labels           []string   `json:"labels,omitempty" form:"labels"`
	Slug             *string    `json:"slug,omitempty" form:"slug"`
	UTM              *UTMParams `json:"utm,omitempty"`
}
