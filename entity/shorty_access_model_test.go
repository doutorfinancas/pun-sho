package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestShortyAccess_ConvertMeta(t *testing.T) {
	type fields struct {
		ID              uuid.UUID
		CreatedAt       *time.Time
		ShortyID        uuid.UUID
		Meta            Meta
		UserAgent       string
		IPAddress       string
		Extra           string
		OperatingSystem string
		Browser         string
		Status          string
	}
	type args struct {
		a map[string][]string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Meta
	}{
		{
			name: "TestMetaCreation",
			fields: fields{
				ID:              uuid.UUID{},
				ShortyID:        uuid.UUID{},
				Meta:            Meta{},
				UserAgent:       "",
				IPAddress:       "130.0.0.0",
				Extra:           "EXTRA",
				OperatingSystem: "WINDOWS",
				Browser:         "CHROME",
				Status:          "PASS",
			},
			args: args{
				a: map[string][]string{
					"first": {
						"arg1",
						"arg2",
					},
					"second": {
						"arg3",
					},
				},
			},
			want: Meta{
				M: []MetaValues{
					{
						Name: "first",
						Values: []string{
							"arg1",
							"arg2",
						},
					},
					{
						Name: "second",
						Values: []string{
							"arg3",
						},
					},
				},
			},
		},
		{
			name: "TestMetaUpdate",
			fields: fields{
				ID:       uuid.UUID{},
				ShortyID: uuid.UUID{},
				Meta: Meta{
					M: []MetaValues{
						{
							Name: "Already",
							Values: []string{
								"valueHere",
							},
						},
					},
				},
				UserAgent:       "",
				IPAddress:       "130.0.0.0",
				Extra:           "EXTRA",
				OperatingSystem: "WINDOWS",
				Browser:         "CHROME",
				Status:          "PASS",
			},
			args: args{
				a: map[string][]string{
					"ze": {
						"braco",
					},
					"toni": {
						"perna",
						"cabeca",
					},
				},
			},
			want: Meta{
				M: []MetaValues{
					{
						Name: "toni",
						Values: []string{
							"perna",
							"cabeca",
						},
					},
					{
						Name: "ze",
						Values: []string{
							"braco",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				sh := ShortyAccess{
					ShortyID:        tt.fields.ShortyID,
					Meta:            tt.fields.Meta,
					UserAgent:       tt.fields.UserAgent,
					IPAddress:       tt.fields.IPAddress,
					Extra:           tt.fields.Extra,
					OperatingSystem: tt.fields.OperatingSystem,
					Browser:         tt.fields.Browser,
					Status:          tt.fields.Status,
				}
				got := sh.ConvertMeta(tt.args.a)
				assert.ElementsMatch(t, got.M, tt.want.M, "ConvertMeta() failed")
			},
		)
	}
}
