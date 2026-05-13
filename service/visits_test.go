package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDateRange(t *testing.T) {
	tests := []struct {
		name      string
		fromStr   string
		untilStr  string
		wantFrom  *time.Time
		wantUntil *time.Time
		wantErr   bool
	}{
		{
			name:     "both empty returns nil without error",
			fromStr:  "",
			untilStr: "",
		},
		{
			name:     "valid dates parsed successfully",
			fromStr:  "2026-01-01",
			untilStr: "2026-03-31",
			wantFrom: datePtr(2026, 1, 1, 0, 0, 0),
			wantUntil: datePtr(2026, 3, 31, 23, 59, 59),
		},
		{
			name:    "invalid from date returns error",
			fromStr: "01-01-2026",
			wantErr: true,
		},
		{
			name:     "invalid until date returns error",
			fromStr:  "",
			untilStr: "abc",
			wantErr:  true,
		},
		{
			name:    "from after until returns error",
			fromStr: "2026-05-01",
			untilStr: "2026-01-01",
			wantErr: true,
		},
		{
			name:     "same day is valid",
			fromStr:  "2026-03-15",
			untilStr: "2026-03-15",
			wantFrom: datePtr(2026, 3, 15, 0, 0, 0),
			wantUntil: datePtr(2026, 3, 15, 23, 59, 59),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, until, err := ParseDateRange(tt.fromStr, tt.untilStr)

			if tt.wantErr {
				require.Error(t, err)
				var valErr *ValidationError
				assert.True(t, errors.As(err, &valErr), "expected ValidationError")
				return
			}

			require.NoError(t, err)

			if tt.wantFrom == nil {
				assert.Nil(t, from)
			} else {
				require.NotNil(t, from)
				assert.Equal(t, *tt.wantFrom, *from)
			}

			if tt.wantUntil == nil {
				assert.Nil(t, until)
			} else {
				require.NotNil(t, until)
				assert.Equal(t, *tt.wantUntil, *until)
			}
		})
	}
}

func TestParseGranularity(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "empty string defaults to auto", input: "", want: GranularityAuto},
		{name: "auto is valid", input: "auto", want: GranularityAuto},
		{name: "day is valid", input: "day", want: GranularityDay},
		{name: "month is invalid", input: "month", wantErr: true},
		{name: "week is invalid", input: "week", wantErr: true},
		{name: "year is invalid", input: "year", wantErr: true},
		{name: "arbitrary string is invalid", input: "quarterly", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGranularity(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				var valErr *ValidationError
				assert.True(t, errors.As(err, &valErr), "expected ValidationError")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveGranularity(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		days        int
		granularity string
		want        string
	}{
		{name: "1 day → day", days: 1, granularity: GranularityAuto, want: GranularityDay},
		{name: "12 days → week (lower limit)", days: 12, granularity: GranularityAuto, want: GranularityWeek},
		{name: "11 days → day (upper limit)", days: 11, granularity: GranularityAuto, want: GranularityDay},
		{name: "60 days → week (upper limit)", days: 60, granularity: GranularityAuto, want: GranularityWeek},
		{name: "61 days → month (lower limit)", days: 61, granularity: GranularityAuto, want: GranularityMonth},
		{name: "365 days → month (upper limit)", days: 365, granularity: GranularityAuto, want: GranularityMonth},
		{name: "366 days → year (lower limit)", days: 366, granularity: GranularityAuto, want: GranularityYear},
		{name: "730 days → year", days: 730, granularity: GranularityAuto, want: GranularityYear},
		{name: "granularity=day forced for any range", days: 400, granularity: GranularityDay, want: GranularityDay},
		{name: "granularity=day forced for 1 day", days: 1, granularity: GranularityDay, want: GranularityDay},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			until := base.AddDate(0, 0, tt.days)
			got := ResolveGranularity(base, until, tt.granularity)
			assert.Equal(t, tt.want, got)
		})
	}
}

func datePtr(year, month, day, hour, minute, sec int) *time.Time {
	t := time.Date(year, time.Month(month), day, hour, minute, sec, 0, time.UTC)
	return &t
}
