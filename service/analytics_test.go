package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/database"
	mockDB "github.com/doutorfinancas/pun-sho/test"
)

func TestAnalyticsService_ClicksOverTime_EmptyResult(t *testing.T) {
	_, gormDB := mockDB.NewMockDB()
	log, _ := zap.NewDevelopment()
	db := database.NewDatabase(gormDB)

	svc := NewAnalyticsService(log, db)

	from := time.Now().AddDate(0, 0, -30)
	until := time.Now()

	// Without setting up mock expectations, this will return empty
	// This tests that the service handles empty results gracefully
	points := svc.ClicksOverTime(nil, from, until, "day")
	assert.NotNil(t, points)
}

func TestAnalyticsService_ClicksOverTime_WithShortyID(t *testing.T) {
	_, gormDB := mockDB.NewMockDB()
	log, _ := zap.NewDevelopment()
	db := database.NewDatabase(gormDB)

	svc := NewAnalyticsService(log, db)

	id := uuid.New()
	from := time.Now().AddDate(0, 0, -30)
	until := time.Now()

	points := svc.ClicksOverTime(&id, from, until, "week")
	assert.NotNil(t, points)
}

func TestTimeseriesPoint_Structure(t *testing.T) {
	p := TimeseriesPoint{Label: "2024-01-01", Value: 42}
	assert.Equal(t, "2024-01-01", p.Label)
	assert.Equal(t, 42, p.Value)
}

func TestBreakdownItem_Structure(t *testing.T) {
	item := BreakdownItem{Name: "Chrome 120", Count: 100}
	assert.Equal(t, "Chrome 120", item.Name)
	assert.Equal(t, 100, item.Count)
}

func TestGlobalStats_Structure(t *testing.T) {
	stats := GlobalStats{
		TotalLinks:   50,
		TotalClicks:  1000,
		ActiveLinks:  45,
		ExpiredLinks: 5,
	}
	assert.Equal(t, int64(50), stats.TotalLinks)
	assert.Equal(t, int64(1000), stats.TotalClicks)
}

func TestLabelRankItem_Structure(t *testing.T) {
	item := LabelRankItem{Label: "marketing", Count: 500}
	assert.Equal(t, "marketing", item.Label)
	assert.Equal(t, 500, item.Count)
}
