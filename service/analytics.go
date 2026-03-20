package service

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/database"
)

type TimeseriesPoint struct {
	Label string `json:"label" gorm:"label"`
	Value int    `json:"value" gorm:"value"`
}

type BreakdownItem struct {
	Name  string `json:"name" gorm:"name"`
	Count int    `json:"count" gorm:"count"`
}

type LocationItem struct {
	Country string `json:"country" gorm:"country"`
	City    string `json:"city" gorm:"city"`
	Count   int    `json:"count" gorm:"count"`
}

type LabelRankItem struct {
	Label string `json:"label" gorm:"label"`
	Count int    `json:"count" gorm:"count"`
}

type GlobalStats struct {
	TotalLinks   int64 `json:"total_links"`
	TotalClicks  int64 `json:"total_clicks"`
	ActiveLinks  int64 `json:"active_links"`
	ExpiredLinks int64 `json:"expired_links"`
}

type AnalyticsService struct {
	log *zap.Logger
	db  *database.Database
}

func NewAnalyticsService(log *zap.Logger, db *database.Database) *AnalyticsService {
	return &AnalyticsService{
		log: log,
		db:  db,
	}
}

func (s *AnalyticsService) ClicksOverTime(shortyID *uuid.UUID, from, until time.Time, granularity string) []TimeseriesPoint {
	truncFunc := "day"
	switch granularity {
	case "week":
		truncFunc = "week"
	case "month":
		truncFunc = "month"
	}

	query := `
		SELECT date_trunc(?, created_at)::date::text as label, COUNT(*) as value
		FROM shorty_accesses
		WHERE created_at BETWEEN ? AND ?
		AND status = 'redirected'`

	args := []interface{}{truncFunc, from, until}

	if shortyID != nil {
		query += " AND shorty_id = ?"
		args = append(args, *shortyID)
	}

	query += " GROUP BY label ORDER BY label"

	var points []TimeseriesPoint
	s.db.Orm.Raw(query, args...).Scan(&points)
	if points == nil {
		points = []TimeseriesPoint{}
	}

	return points
}

func (s *AnalyticsService) BrowserBreakdown(shortyID *uuid.UUID, from, until time.Time) []BreakdownItem {
	var items []BreakdownItem

	query := `
		SELECT browser as name, COUNT(*) as count
		FROM shorty_accesses
		WHERE created_at BETWEEN ? AND ?
		AND status = 'redirected'`

	args := []interface{}{from, until}

	if shortyID != nil {
		query += " AND shorty_id = ?"
		args = append(args, *shortyID)
	}

	query += " GROUP BY browser ORDER BY count DESC LIMIT 10"

	s.db.Orm.Raw(query, args...).Scan(&items)
	if items == nil {
		items = []BreakdownItem{}
	}

	return items
}

func (s *AnalyticsService) OSBreakdown(shortyID *uuid.UUID, from, until time.Time) []BreakdownItem {
	var items []BreakdownItem

	query := `
		SELECT operating_system as name, COUNT(*) as count
		FROM shorty_accesses
		WHERE created_at BETWEEN ? AND ?
		AND status = 'redirected'`

	args := []interface{}{from, until}

	if shortyID != nil {
		query += " AND shorty_id = ?"
		args = append(args, *shortyID)
	}

	query += " GROUP BY operating_system ORDER BY count DESC LIMIT 10"

	s.db.Orm.Raw(query, args...).Scan(&items)
	if items == nil {
		items = []BreakdownItem{}
	}

	return items
}

func (s *AnalyticsService) TopReferrers(shortyID *uuid.UUID, from, until time.Time, limit int) []BreakdownItem {
	if limit == 0 {
		limit = 10
	}

	var items []BreakdownItem

	query := `
		SELECT COALESCE(
			(meta->'meta_collection'->0->>'values')::text,
			'Direct'
		) as name, COUNT(*) as count
		FROM shorty_accesses
		WHERE created_at BETWEEN ? AND ?
		AND status = 'redirected'`

	args := []interface{}{from, until}

	if shortyID != nil {
		query += " AND shorty_id = ?"
		args = append(args, *shortyID)
	}

	query += " GROUP BY name ORDER BY count DESC LIMIT ?"
	args = append(args, limit)

	s.db.Orm.Raw(query, args...).Scan(&items)
	if items == nil {
		items = []BreakdownItem{}
	}

	return items
}

func (s *AnalyticsService) LocationBreakdown(shortyID *uuid.UUID, from, until time.Time, limit int) []LocationItem {
	if limit == 0 {
		limit = 10
	}

	var items []LocationItem

	query := `
		SELECT COALESCE(country, 'Unknown') as country, COALESCE(city, 'Unknown') as city, COUNT(*) as count
		FROM shorty_accesses
		WHERE created_at BETWEEN ? AND ?
		AND status = 'redirected'`

	args := []interface{}{from, until}

	if shortyID != nil {
		query += " AND shorty_id = ?"
		args = append(args, *shortyID)
	}

	query += " GROUP BY country, city ORDER BY count DESC LIMIT ?"
	args = append(args, limit)

	s.db.Orm.Raw(query, args...).Scan(&items)
	if items == nil {
		items = []LocationItem{}
	}

	return items
}

func (s *AnalyticsService) LabelRanking(from, until time.Time, limit int) []LabelRankItem {
	if limit == 0 {
		limit = 10
	}

	var items []LabelRankItem

	s.db.Orm.Raw(`
		SELECT unnest(s.labels) as label, COUNT(sa.id) as count
		FROM shorties s
		INNER JOIN shorty_accesses sa ON s.id = sa.shorty_id
		WHERE sa.created_at BETWEEN ? AND ?
		AND sa.status = 'redirected'
		AND s.labels IS NOT NULL
		GROUP BY label
		ORDER BY count DESC
		LIMIT ?`, from, until, limit).Scan(&items)

	if items == nil {
		items = []LabelRankItem{}
	}

	return items
}

func (s *AnalyticsService) GlobalSummary(from, until time.Time) GlobalStats {
	var stats GlobalStats

	s.db.Orm.Raw(`SELECT COUNT(*) FROM shorties WHERE deleted_at IS NULL`).Scan(&stats.TotalLinks)

	s.db.Orm.Raw(`
		SELECT COUNT(*) FROM shorty_accesses
		WHERE created_at BETWEEN ? AND ?
		AND status = 'redirected'`, from, until).Scan(&stats.TotalClicks)

	s.db.Orm.Raw(`
		SELECT COUNT(*) FROM shorties
		WHERE deleted_at IS NULL
		AND (ttl IS NULL OR ttl > NOW())`).Scan(&stats.ActiveLinks)

	s.db.Orm.Raw(`
		SELECT COUNT(*) FROM shorties
		WHERE deleted_at IS NULL
		AND ttl IS NOT NULL AND ttl <= NOW()`).Scan(&stats.ExpiredLinks)

	return stats
}

