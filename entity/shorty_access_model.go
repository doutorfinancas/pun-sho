package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type ShortyAccess struct {
	BaseModel
	ShortyID        uuid.UUID `json:"-" gorm:"column:shorty_id"`
	Meta            Meta      `json:"meta" gorm:"column:meta; type:JSONB"`
	UserAgent       string    `json:"user_agent" gorm:"column:user_agent"`
	IPAddress       string    `json:"ip" gorm:"column:ip_address"`
	Extra           string    `json:"extra" gorm:"column:extra"`
	OperatingSystem string    `json:"os" gorm:"column:operating_system"`
	Browser         string    `json:"browser" gorm:"column:browser"`
	Status          string    `json:"status" gorm:"column:status"`
}

func (ShortyAccess) TableName() string {
	return "shorty_accesses"
}

type Meta struct {
	M []MetaValues `json:"meta_collection"`
}

type MetaValues struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

func (ShortyAccess) ConvertMeta(a map[string][]string) Meta {
	var tmp []MetaValues

	for k, v := range a {
		tmp = append(tmp, MetaValues{Name: k, Values: v})
	}

	return Meta{M: tmp}
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *Meta) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := Meta{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

// Value return json value, implement driver.Valuer interface
func (j *Meta) Value() (driver.Value, error) {
	if len(j.M) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}
