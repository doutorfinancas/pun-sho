package entity

import (
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
