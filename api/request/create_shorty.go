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

type QRCode struct {
	Create      bool   `json:"create"`
	Width       int    `json:"width"`
	BorderWidth int    `json:"border_width"`
	FgColor     string `json:"foreground_color"`
	BgColor     string `json:"background_color"`
	Shape       string `json:"shape"`
}
