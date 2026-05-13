package service

import (
	"fmt"
	"time"
)

const (
	GranularityAuto  = "auto"
	GranularityDay   = "day"
	GranularityWeek  = "week"
	GranularityMonth = "month"
	GranularityYear  = "year"
)

// ValidationError is returned for invalid input parameters.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type VisitStatsPeriod struct {
	Date   string `json:"date"`
	Visits int64  `json:"visits"`
}

// VisitStatsResponse is a map keyed by period start date (YYYY-MM-DD).
// All periods in the requested range are present, including those with zero visits.
type VisitStatsResponse map[string]VisitStatsPeriod

// ParseDateRange parses optional from/until strings (YYYY-MM-DD format).
// Returns (nil, nil, nil) when both strings are empty.
// The until time is set to end-of-day (23:59:59) to include the full day.
func ParseDateRange(fromStr, untilStr string) (from, until *time.Time, err error) {
	if fromStr == "" && untilStr == "" {
		return nil, nil, nil
	}

	if fromStr != "" {
		t, parseErr := time.Parse(time.DateOnly, fromStr)
		if parseErr != nil {
			return nil, nil, &ValidationError{
				Message: fmt.Sprintf("invalid from date %q: must be YYYY-MM-DD", fromStr),
			}
		}
		from = &t
	}

	if untilStr != "" {
		t, parseErr := time.Parse(time.DateOnly, untilStr)
		if parseErr != nil {
			return nil, nil, &ValidationError{
				Message: fmt.Sprintf("invalid until date %q: must be YYYY-MM-DD", untilStr),
			}
		}
		endOfDay := t.Add(24*time.Hour - time.Second)
		until = &endOfDay
	}

	if from != nil && until != nil && from.After(*until) {
		return nil, nil, &ValidationError{Message: "from date must not be after until date"}
	}

	return from, until, nil
}

// ParseGranularity validates the granularity string. Valid values: "", "auto", "day".
func ParseGranularity(granularityStr string) (string, error) {
	switch granularityStr {
	case "", GranularityAuto:
		return GranularityAuto, nil
	case GranularityDay:
		return GranularityDay, nil
	default:
		return "", &ValidationError{
			Message: fmt.Sprintf("invalid granularity %q: must be 'auto' or 'day'", granularityStr),
		}
	}
}

// ResolveGranularity returns the effective grouping granularity for the given range.
// If granularity is GranularityDay it is returned unchanged.
// Otherwise the auto rule applies: <12d→day, ≤60d→week, ≤365d→month, >365d→year.
func ResolveGranularity(from, until time.Time, granularity string) string {
	if granularity == GranularityDay {
		return GranularityDay
	}
	days := int(until.Sub(from).Hours() / 24)
	switch {
	case days < 12:
		return GranularityDay
	case days <= 60:
		return GranularityWeek
	case days <= 365:
		return GranularityMonth
	default:
		return GranularityYear
	}
}

// generatePeriods returns all period start dates (YYYY-MM-DD) between from and until
// for the given granularity. Periods with no visits are included (for zero-filling).
func generatePeriods(from, until time.Time, granularity string) []string {
	var periods []string

	switch granularity {
	case GranularityDay:
		cur := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
		for !cur.After(until) {
			periods = append(periods, cur.Format(time.DateOnly))
			cur = cur.AddDate(0, 0, 1)
		}

	case GranularityWeek:
		// Find the Monday of from's week
		wd := int(from.Weekday())
		if wd == 0 {
			wd = 7 // Sunday → 7 so offset becomes -6
		}
		monday := from.AddDate(0, 0, -(wd - 1))
		cur := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
		for !cur.After(until) {
			periods = append(periods, cur.Format(time.DateOnly))
			cur = cur.AddDate(0, 0, 7)
		}

	case GranularityMonth:
		cur := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)
		for !cur.After(until) {
			periods = append(periods, cur.Format(time.DateOnly))
			cur = cur.AddDate(0, 1, 0)
		}

	case GranularityYear:
		cur := time.Date(from.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		for !cur.After(until) {
			periods = append(periods, cur.Format(time.DateOnly))
			cur = cur.AddDate(1, 0, 0)
		}
	}

	return periods
}
