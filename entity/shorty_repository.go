package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/database"
)

type ShortyRepository struct {
	database.Repository
}

func NewShortyRepository(db *database.Database, log *zap.Logger) *ShortyRepository {
	return &ShortyRepository{
		database.Repository{
			Database: db,
			Logger:   log,
		},
	}
}

func (r *ShortyRepository) List(limit, offset int) ([]*Shorty, error) {
	rows := make([]*Shorty, 0)

	if err := r.Database.FetchPage(
		&Shorty{},
		limit,
		offset,
		&rows,
	); err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *ShortyRepository) ListWithAccessData(
	listWithQRCode bool,
	labels []string,
	status string,
	from, to *time.Time,
	limit, offset int,
) ([]*ShortyForList, error) {
	rows := make([]*ShortyForList, 0)

	fields := []string{
		"s.id",
		"s.created_at",
		"s.deleted_at",
		"s.public_id",
		"s.link",
		"s.ttl",
		"s.redirection_limit",
		"s.labels",
	}

	if listWithQRCode {
		fields = append(fields, "qr_code")
	}

	fieldlist := strings.Join(fields, ", ")

	baseQuery := fmt.Sprintf(
		`SELECT %s, count(sa.id) as visits, COALESCE(sum(CASE WHEN sa.status = 'redirected' THEN 1 ELSE 0 END), 0) as redirects FROM shorties s
    LEFT JOIN shorty_accesses sa
        ON s.id = sa.shorty_id`, fieldlist)

	var conditions []string
	var args []interface{}

	if len(labels) > 0 {
		conditions = append(conditions, "s.labels && ?")
		args = append(args, fmt.Sprintf("{%s}", strings.Join(labels, ",")))
	}

	switch status {
	case "active":
		conditions = append(conditions, "s.deleted_at IS NULL AND (s.ttl IS NULL OR s.ttl > NOW())")
	case "expired":
		conditions = append(conditions, "s.deleted_at IS NULL AND s.ttl IS NOT NULL AND s.ttl <= NOW() AND s.ttl > '0002-01-01'")
	case "deleted":
		conditions = append(conditions, "s.deleted_at IS NOT NULL")
	}

	if from != nil {
		conditions = append(conditions, "s.created_at >= ?")
		args = append(args, *from)
	}

	if to != nil {
		conditions = append(conditions, "s.created_at <= ?")
		args = append(args, *to)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	groupByClause := " GROUP BY s.id, s.created_at, s.deleted_at, s.public_id, s.link, s.qr_code, s.ttl, s.redirection_limit, s.labels LIMIT ? OFFSET ?"

	query := baseQuery + whereClause + groupByClause
	args = append(args, limit, offset)

	if err := r.Database.Orm.Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *ShortyRepository) DistinctLabels() ([]string, error) {
	var labels []string
	err := r.Database.Orm.Raw(
		`SELECT DISTINCT unnest(labels) AS label FROM shorties WHERE deleted_at IS NULL ORDER BY 1`,
	).Scan(&labels).Error
	if err != nil {
		return nil, err
	}
	return labels, nil
}

func (r *ShortyRepository) ExistsByPublicID(publicID string) (bool, error) {
	var count int64
	if err := r.Database.Orm.
		Model(&Shorty{}).
		Where("public_id = ?", publicID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ShortyRepository) Delete(id uuid.UUID) error {
	m := Shorty{ID: id}
	return r.Database.Orm.Model(m).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).
		Error
}
