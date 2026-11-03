package entity

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/database"
)

type ShortyAccessRepository struct {
	database.Repository
}

func NewShortyAccessRepository(db *database.Database, log *zap.Logger) *ShortyAccessRepository {
	return &ShortyAccessRepository{
		database.Repository{
			Database: db,
			Logger:   log,
		},
	}
}

// CountByShortyID returns the total visit count and the redirected count for a
// shorty in one round-trip — much cheaper than loading every access row to
// take len() of the slice.
func (r *ShortyAccessRepository) CountByShortyID(id uuid.UUID) (visits, redirects int64, err error) {
	var row struct {
		Visits    int64 `gorm:"visits"`
		Redirects int64 `gorm:"redirects"`
	}
	err = r.Database.Orm.Raw(`
		SELECT COUNT(*) AS visits,
		       COALESCE(SUM(CASE WHEN status = 'redirected' THEN 1 ELSE 0 END), 0) AS redirects
		FROM shorty_accesses
		WHERE shorty_id = ?`, id).Scan(&row).Error
	return row.Visits, row.Redirects, err
}

// CountRedirectsByShortyID counts only redirected accesses, served by idx_shorty_accesses_shorty_redirected.
func (r *ShortyAccessRepository) CountRedirectsByShortyID(id uuid.UUID) (int64, error) {
	var count int64
	err := r.Database.Orm.
		Model(&ShortyAccess{}).
		Where("shorty_id = ? AND status = ?", id, "redirected").
		Count(&count).Error
	return count, err
}

// ShortyVisitCount holds lifetime totals for one link.
type ShortyVisitCount struct {
	ID        uuid.UUID `json:"id" gorm:"column:id"`
	Visits    int64     `json:"visits" gorm:"column:visits"`
	Redirects int64     `json:"redirects" gorm:"column:redirects"`
}

// VisitCountsSince returns lifetime totals only for links visited at or after since.
// The CTE and per-link lookups keep the cost proportional to recently visited links, not to the table or label size.
func (r *ShortyAccessRepository) VisitCountsSince(
	ctx context.Context, labels []string, since time.Time,
) ([]ShortyVisitCount, error) {
	labelFilter := ""
	args := []interface{}{since}
	if len(labels) > 0 {
		labelFilter = "WHERE (SELECT s.labels FROM shorties s WHERE s.id = r.shorty_id) && ?"
		args = append(args, labelsArrayLiteral(labels))
	}

	rows := make([]ShortyVisitCount, 0)
	err := r.Database.Orm.WithContext(ctx).Raw(fmt.Sprintf(`
		WITH recent AS MATERIALIZED (
			SELECT DISTINCT shorty_id FROM shorty_accesses WHERE created_at >= ?
		)
		SELECT r.shorty_id AS id, c.visits, c.redirects
		FROM recent r
		CROSS JOIN LATERAL (
			SELECT COUNT(*) AS visits,
			       COALESCE(SUM(CASE WHEN sa.status = 'redirected' THEN 1 ELSE 0 END), 0) AS redirects
			FROM shorty_accesses sa
			WHERE sa.shorty_id = r.shorty_id
		) c
		%s
		ORDER BY r.shorty_id`, labelFilter), args...).Scan(&rows).Error

	return rows, err
}

// CountByShortyIDAndDateRange is the date-bounded equivalent of CountByShortyID.
func (r *ShortyAccessRepository) CountByShortyIDAndDateRange(
	id uuid.UUID, from, until *time.Time,
) (visits, redirects int64, err error) {
	var row struct {
		Visits    int64 `gorm:"visits"`
		Redirects int64 `gorm:"redirects"`
	}
	err = r.Database.Orm.Raw(`
		SELECT COUNT(*) AS visits,
		       COALESCE(SUM(CASE WHEN status = 'redirected' THEN 1 ELSE 0 END), 0) AS redirects
		FROM shorty_accesses
		WHERE shorty_id = ? AND created_at BETWEEN ? AND ?`, id, from, until).Scan(&row).Error
	return row.Visits, row.Redirects, err
}

func (r *ShortyAccessRepository) ListByShortyUUID(id uuid.UUID, limit, offset int) ([]*ShortyAccess, error) {
	rows := make([]*ShortyAccess, 0)

	if err := r.Database.Orm.
		Model(
			&ShortyAccess{},
		).
		Select("shorty_accesses.*").
		Limit(limit).
		Offset(offset).
		Where("shorty_accesses.shorty_id = ?", id).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
