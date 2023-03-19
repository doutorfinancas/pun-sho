package request

import (
	"time"
)

type CreateShorty struct {
	Link string     `json:"link"`
	TTL  *time.Time `json:"TTL"`
}
