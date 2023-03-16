package test

import "time"

// TestsModel Model for testing gorm database
type TestsModel struct {
	ID         int64     `gorm:"column:id"`
	TestInt    int64     `gorm:"column:test_int"`
	TestString string    `gorm:"column:test_string"`
	TestBool   bool      `gorm:"column:test_bool"`
	TestFloat  float64   `gorm:"column:test_float"`
	TestTime   time.Time `gorm:"column:test_time"`
}

func (TestsModel) TableName() string {
	return "tests"
}
