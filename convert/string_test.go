package convert

import (
	"crypto"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestToStringable struct {
	Value string
}

func (s TestToStringable) ToString() string {
	return s.Value
}

func TestToString(t *testing.T) {
	var ns *string
	ptn := &ns
	ptn = nil
	s := "test"

	i := int(10)
	i64 := int64(10)
	f := 10.1
	f32 := float32(10.1)
	f64 := float64(10.1)
	bf := big.NewFloat(10.1)

	tn := time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)

	err := errors.New("this is an error")

	toStringable := TestToStringable{Value: "this is Stringable"}

	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test nil value", args{nil}, ""},
		{"test nil pointer", args{ptn}, ""},
		{"test nil string", args{ns}, ""},
		{"test nil string pointer", args{&ns}, ""},
		{"test empty string", args{""}, ""},
		{"test string to string", args{s}, "test"},
		{"test *string to string", args{&s}, "test"},
		{"test int to string", args{i}, "10"},
		{"test *int to string", args{&i}, "10"},
		{"test int64 to string", args{i64}, "10"},
		{"test *int64 to string", args{&i64}, "10"},
		{"test float to string", args{f}, "10.1"},
		{"test *float to string", args{&f}, "10.1"},
		{"test float32 to string", args{f32}, "10.1"},
		{"test *float32 to string", args{&f32}, "10.1"},
		{"test float64 to string", args{f64}, "10.1"},
		{"test *float64 to string", args{&f64}, "10.1"},
		{"test *bigFloat to string", args{bf}, "10.1"},
		{"test time.Time to string", args{tn}, "2019-01-01T00:00:00Z"},
		{"test *time.Time to string", args{&tn}, "2019-01-01T00:00:00Z"},
		{"test unknown type to string", args{t}, ""},
		{"test *unknown type to string", args{&t}, ""},
		{"test error to string", args{err}, "this is an error"},
		{"test ToStringable interface to string", args{toStringable}, "this is Stringable"},
		{"test *crypto.SHA256 to string", args{crypto.SHA256}, "SHA-256"},
		{"test *crypto.SHA512 to string", args{crypto.SHA512}, "SHA-512"},
		{"test *crypto.UNKNOWN to string", args{crypto.MD5}, "MD5"},
		{"test []string to string", args{[]string{"this", "is", "it"}}, "this; is; it"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.args.v); got != tt.want {
				t.Errorf("ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToStringNil(t *testing.T) {
	var ns *string
	ptn := &ns
	ptn = nil
	s := "test"
	es := ""

	i := int(10)
	i64 := int64(10)
	f := 10.1
	f32 := float32(10.1)
	f64 := float64(10.1)
	bf := big.NewFloat(10.1)

	tn := time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)

	err := errors.New("test")

	toStringable := TestToStringable{Value: "test"}

	nsTest := "test"
	ns10 := "10"
	ns10dot1 := "10.1"
	nsDate := "2019-01-01T00:00:00Z"

	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want *string
	}{
		{"test nil value", args{nil}, nil},
		{"test nil pointer", args{ptn}, nil},
		{"test nil string", args{ns}, nil},
		{"test nil string pointer", args{&ns}, nil},
		{"test empty string", args{""}, &es},
		{"test string to string", args{s}, &nsTest},
		{"test *string to string", args{&s}, &nsTest},
		{"test int to string", args{i}, &ns10},
		{"test *int to string", args{&i}, &ns10},
		{"test int64 to string", args{i64}, &ns10},
		{"test *int64 to string", args{&i64}, &ns10},
		{"test float to string", args{f}, &ns10dot1},
		{"test *float to string", args{&f}, &ns10dot1},
		{"test float32 to string", args{f32}, &ns10dot1},
		{"test *float32 to string", args{&f32}, &ns10dot1},
		{"test float64 to string", args{f64}, &ns10dot1},
		{"test *float64 to string", args{&f64}, &ns10dot1},
		{"test *bigFloat to string", args{bf}, &ns10dot1},
		{"test time.Time to string", args{tn}, &nsDate},
		{"test *time.Time to string", args{&tn}, &nsDate},
		{"test unknown type to string", args{t}, nil},
		{"test *unknown type to string", args{&t}, nil},
		{"test error to string", args{err}, &nsTest},
		{"test ToStringable interface to string", args{toStringable}, &nsTest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToStringNil(tt.args.v); !assert.ObjectsAreEqual(tt.want, got) {
				t.Errorf("ToStringNil() = %v, want %v", got, tt.want)
			}
		})
	}
}
