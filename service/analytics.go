package service

import (
	"sync"
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

	summaryMu     sync.Mutex
	summaryCache  GlobalStats
	summaryExpiry time.Time
	summaryKey    string
}

// summaryCacheTTL is how long the dashboard's GlobalSummary counts are reused
// across requests. Counts move slowly relative to dashboard refresh cadence,
// so a short TTL gives near-instant responses without showing stale data.
const summaryCacheTTL = 15 * time.Second

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
	if err := s.db.Orm.Raw(query, args...).Scan(&points).Error; err != nil {
		s.log.Error("ClicksOverTime query failed", zap.Error(err))
	}
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

	if err := s.db.Orm.Raw(query, args...).Scan(&items).Error; err != nil {
		s.log.Error("Analytics query failed", zap.Error(err))
	}
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

	if err := s.db.Orm.Raw(query, args...).Scan(&items).Error; err != nil {
		s.log.Error("Analytics query failed", zap.Error(err))
	}
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

	if err := s.db.Orm.Raw(query, args...).Scan(&items).Error; err != nil {
		s.log.Error("Analytics query failed", zap.Error(err))
	}
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

	if err := s.db.Orm.Raw(query, args...).Scan(&items).Error; err != nil {
		s.log.Error("LocationBreakdown query failed", zap.Error(err))
	}
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

	if err := s.db.Orm.Raw(`
		SELECT unnest(s.labels) as label, COUNT(sa.id) as count
		FROM shorties s
		INNER JOIN shorty_accesses sa ON s.id = sa.shorty_id
		WHERE sa.created_at BETWEEN ? AND ?
		AND sa.status = 'redirected'
		AND s.labels IS NOT NULL
		GROUP BY label
		ORDER BY count DESC
		LIMIT ?`, from, until, limit).Scan(&items).Error; err != nil {
		s.log.Error("LabelRanking query failed", zap.Error(err))
	}

	if items == nil {
		items = []LabelRankItem{}
	}

	return items
}

func (s *AnalyticsService) GlobalSummary(from, until time.Time) GlobalStats {
	// Cache hit: serve immediately. The window (from/until) is part of the key
	// so dashboard variations don't share counts.
	cacheKey := from.UTC().Format(time.RFC3339) + "|" + until.UTC().Format(time.RFC3339)
	s.summaryMu.Lock()
	if s.summaryKey == cacheKey && time.Now().Before(s.summaryExpiry) {
		cached := s.summaryCache
		s.summaryMu.Unlock()
		return cached
	}
	s.summaryMu.Unlock()

	// Counts are independent — run them in parallel so the dashboard doesn't
	// hang on the slowest one (on large tables each COUNT can take seconds).
	type countResult struct {
		Count int64 `gorm:"count"`
	}

	runCount := func(query string, args ...interface{}) int64 {
		var r countResult
		if err := s.db.Orm.Raw(query, args...).Scan(&r).Error; err != nil {
			s.log.Error("GlobalSummary count query failed", zap.String("query", query), zap.Error(err))
		}
		return r.Count
	}

	var (
		stats          GlobalStats
		activeNoTTL    int64
		activeFutureTTL int64
		wg             sync.WaitGroup
	)

	// The active-links count was previously `(ttl IS NULL OR ttl > NOW())`,
	// which Postgres can't satisfy with a single B-tree scan because of the
	// IS NULL branch. Splitting into two halves lets each hit a partial index
	// (idx_shorties_active_no_ttl and idx_shorties_active_ttl).
	wg.Add(5)
	go func() {
		defer wg.Done()
		stats.TotalLinks = runCount(`SELECT COUNT(*) AS count FROM shorties WHERE deleted_at IS NULL`)
	}()
	go func() {
		defer wg.Done()
		stats.TotalClicks = runCount(`
			SELECT COUNT(*) AS count FROM shorty_accesses
			WHERE created_at BETWEEN ? AND ?
			AND status = 'redirected'`, from, until)
	}()
	go func() {
		defer wg.Done()
		activeNoTTL = runCount(`
			SELECT COUNT(*) AS count FROM shorties
			WHERE deleted_at IS NULL AND ttl IS NULL`)
	}()
	go func() {
		defer wg.Done()
		activeFutureTTL = runCount(`
			SELECT COUNT(*) AS count FROM shorties
			WHERE deleted_at IS NULL AND ttl > NOW()`)
	}()
	go func() {
		defer wg.Done()
		stats.ExpiredLinks = runCount(`
			SELECT COUNT(*) AS count FROM shorties
			WHERE deleted_at IS NULL
			AND ttl IS NOT NULL AND ttl <= NOW()`)
	}()
	wg.Wait()

	stats.ActiveLinks = activeNoTTL + activeFutureTTL

	s.summaryMu.Lock()
	s.summaryCache = stats
	s.summaryKey = cacheKey
	s.summaryExpiry = time.Now().Add(summaryCacheTTL)
	s.summaryMu.Unlock()

	return stats
}

