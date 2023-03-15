package convert

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type stringable interface {
	String() string
}

type Stringable interface {
	ToString() string
}

func ToString(v interface{}) string {
	s := ToStringNil(v)

	if s == nil {
		return ""
	}

	return *s
}

func ToStringNil(v interface{}) *string {
	s := extractStringValue(v)

	if s == nil {
		return nil
	}

	if !utf8.ValidString(*s) {
		// both this and the below method has issues with encoding in name
		_, name, _ := charset.DetermineEncoding([]byte(*s), "")

		if name == "windows-1252" {
			res, _, _ := transform.String(charmap.Windows1252.NewEncoder(), *s)
			s = &res
		} else {
			res, _, _ := transform.String(charmap.ISO8859_1.NewEncoder(), *s)
			s = &res
		}
	}

	return s
}

//nolint:gocyclo
func extractStringValue(v interface{}) *string {
	if v == nil {
		return nil
	}

	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return nil
	}

	switch a := v.(type) {
	case string:
		return &a

	case *string:
		return a

	case int, int64, float64, float32:
		s := fmt.Sprintf("%v", a)

		return &s

	case *int, *int64, *float64, *float32:
		s := fmt.Sprintf("%v", rv.Elem()) //nolint:gocritic

		return &s

	case time.Time:
		s := a.Format(time.RFC3339)

		return &s

	case *time.Time:
		s := a.Format(time.RFC3339)

		return &s

	case byte:
		s := string(a)

		return &s

	case []byte:
		s := string(a)

		return &s

	case []string:
		s := strings.Join(a, "; ")

		return &s
	}

	sa, matches := v.(stringable)
	if matches {
		s := sa.String()

		return &s
	}

	toStringable, matches := v.(Stringable)
	if matches {
		s := toStringable.ToString()

		return &s
	}

	err, matches := v.(error)
	if matches {
		s := err.Error()

		return &s
	}

	return nil
}
