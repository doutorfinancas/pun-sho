package request

import (
	"time"
)

type CreateShorty struct {
	Link             string     `json:"link"`
	TTL              *time.Time `json:"TTL"`
	RedirectionLimit *int       `json:"redirection_limit"`
	QRCode           *QRCode    `json:"qr_code"`
}
