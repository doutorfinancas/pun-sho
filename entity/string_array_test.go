package entity

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringArray_Value(t *testing.T) {
	tests := []struct {
		name    string
		input   StringArray
		want    string
		wantErr bool
	}{
		{
			name:  "empty array",
			input: StringArray{},
			want:  "{}",
		},
		{
			name:  "nil array",
			input: nil,
			want:  "{}",
		},
		{
			name:  "single element",
			input: StringArray{"test"},
			want:  `{"test"}`,
		},
		{
			name:  "multiple elements",
			input: StringArray{"marketing", "campaign", "2024"},
			want:  `{"marketing","campaign","2024"}`,
		},
		{
			name:  "elements with quotes",
			input: StringArray{"test\"quote", "normal"},
			want:  `{"test""quote","normal"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("StringArray.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStringArray_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    StringArray
		wantErr bool
	}{
		{
			name:  "nil input",
			input: nil,
			want:  StringArray{},
		},
		{
			name:  "empty array string",
			input: "{}",
			want:  StringArray{},
		},
		{
			name:  "empty array bytes",
			input: []byte("{}"),
			want:  StringArray{},
		},
		{
			name:  "single element",
			input: "{test}",
			want:  StringArray{"test"},
		},
		{
			name:  "multiple elements",
			input: "{marketing,campaign,2024}",
			want:  StringArray{"marketing", "campaign", "2024"},
		},
		{
			name:  "quoted elements",
			input: `{"marketing","campaign","2024"}`,
			want:  StringArray{"marketing", "campaign", "2024"},
		},
		{
			name:  "elements with spaces",
			input: `{"marketing campaign","2024"}`,
			want:  StringArray{"marketing campaign", "2024"},
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "unsupported type",
			input:   123,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sa StringArray
			err := sa.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringArray.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want, sa)
			}
		})
	}
}

func TestStringArray_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   StringArray
		want    string
		wantErr bool
	}{
		{
			name:  "empty array",
			input: StringArray{},
			want:  "[]",
		},
		{
			name:  "single element",
			input: StringArray{"test"},
			want:  `["test"]`,
		},
		{
			name:  "multiple elements",
			input: StringArray{"marketing", "campaign", "2024"},
			want:  `["marketing","campaign","2024"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringArray.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestStringArray_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    StringArray
		wantErr bool
	}{
		{
			name:  "empty array",
			input: "[]",
			want:  StringArray{},
		},
		{
			name:  "single element",
			input: `["test"]`,
			want:  StringArray{"test"},
		},
		{
			name:  "multiple elements",
			input: `["marketing","campaign","2024"]`,
			want:  StringArray{"marketing", "campaign", "2024"},
		},
		{
			name:    "invalid JSON",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sa StringArray
			err := json.Unmarshal([]byte(tt.input), &sa)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringArray.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want, sa)
			}
		})
	}
}

func TestStringArray_DriverValuer(t *testing.T) {
	// Test that StringArray implements driver.Valuer
	var _ driver.Valuer = StringArray{}
	
	sa := StringArray{"test", "array"}
	val, err := sa.Value()
	assert.NoError(t, err)
	assert.NotNil(t, val)
}

func TestStringArray_SQLScanner(t *testing.T) {
	// Test that StringArray pointer implements sql.Scanner
	var sa StringArray
	assert.Implements(t, (*interface{ Scan(interface{}) error })(nil), &sa)
	
	err := sa.Scan("{test,array}")
	assert.NoError(t, err)
	assert.Equal(t, StringArray{"test", "array"}, sa)
}
