package request

import (
	"time"
)

type UpdateShorty struct {
	Link             string     `json:"link" form:"link"`
	TTL              *time.Time `json:"TTL" form:"ttl"`
	RedirectionLimit *int       `json:"redirection_limit" form:"redirection_limit"`
	Cancel           bool       `json:"cancel" form:"cancel"`
	Labels           []string   `json:"labels,omitempty" form:"labels"`
	UTM              *UTMParams `json:"utm,omitempty"`
}
