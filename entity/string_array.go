package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// StringArray is a generic type for handling PostgreSQL text[] arrays
// It implements the driver.Valuer and sql.Scanner interfaces
type StringArray []string

// Value implements the driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	
	// Escape each string and wrap in quotes
	escaped := make([]string, len(a))
	for i, s := range a {
		escaped[i] = fmt.Sprintf(`"%s"`, strings.ReplaceAll(s, `"`, `""`))
	}
	
	return fmt.Sprintf("{%s}", strings.Join(escaped, ",")), nil
}

// Scan implements the sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}
	
	switch v := value.(type) {
	case string:
		return a.parseString(v)
	case []byte:
		return a.parseString(string(v))
	default:
		return errors.New("cannot scan non-string value into StringArray")
	}
}

// parseString parses PostgreSQL array format string
func (a *StringArray) parseString(s string) error {
	if s == "{}" {
		*a = StringArray{}
		return nil
	}
	
	// Remove surrounding braces
	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return errors.New("invalid array format")
	}
	
	inner := s[1 : len(s)-1]
	if inner == "" {
		*a = StringArray{}
		return nil
	}
	
	// Parse the array elements
	*a = StringArray{}
	var current strings.Builder
	inQuotes := false
	escaped := false
	
	for i, r := range inner {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == '"':
			if !inQuotes && current.Len() > 0 {
				// Handle unquoted elements
				*a = append(*a, current.String())
				current.Reset()
			}
			inQuotes = !inQuotes
		case r == ',' && !inQuotes:
			*a = append(*a, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
		
		// Handle last element
		if i == len(inner)-1 {
			*a = append(*a, current.String())
		}
	}
	
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (a StringArray) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(a))
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (a *StringArray) UnmarshalJSON(data []byte) error {
	var slice []string
	if err := json.Unmarshal(data, &slice); err != nil {
		return err
	}
	*a = StringArray(slice)
	return nil
}
