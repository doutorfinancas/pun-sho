package request

import (
	"time"
)

type UpdateShorty struct {
	Link             string     `json:"link"`
	TTL              *time.Time `json:"TTL"`
	RedirectionLimit *int       `json:"redirection_limit"`
	Cancel           bool       `json:"cancel"`
	Labels           []string   `json:"labels,omitempty"`
}
