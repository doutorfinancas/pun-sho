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

	pageFields := []string{
		"id",
		"created_at",
		"deleted_at",
		"public_id",
		"link",
		"ttl",
		"redirection_limit",
		"labels",
	}
	if listWithQRCode {
		pageFields = append(pageFields, "qr_code")
	}
	pageFieldList := strings.Join(pageFields, ", ")

	selectFields := make([]string, 0, len(pageFields))
	for _, f := range pageFields {
		selectFields = append(selectFields, "p."+f)
	}
	selectFieldList := strings.Join(selectFields, ", ")

	var conditions []string
	var args []interface{}

	if len(labels) > 0 {
		conditions = append(conditions, "labels && ?")
		args = append(args, fmt.Sprintf("{%s}", strings.Join(labels, ",")))
	}

	switch status {
	case "active":
		conditions = append(conditions, "deleted_at IS NULL AND (ttl IS NULL OR ttl > NOW())")
	case "expired":
		conditions = append(conditions, "deleted_at IS NULL AND ttl IS NOT NULL AND ttl <= NOW() AND ttl > '0002-01-01'")
	case "deleted":
		conditions = append(conditions, "deleted_at IS NOT NULL")
	}

	if from != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, *from)
	}
	if to != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, *to)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Pick the page of shorties FIRST (cheap with the new partial indexes),
	// then aggregate accesses for those rows only. Prior implementation did a
	// LEFT JOIN + GROUP BY across the full table, which is O(rows) and
	// catastrophic on a 6M-row table.
	groupBy := make([]string, 0, len(pageFields))
	for _, f := range pageFields {
		groupBy = append(groupBy, "p."+f)
	}

	query := fmt.Sprintf(`
		WITH page AS (
			SELECT %s FROM shorties%s
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		)
		SELECT %s,
			COALESCE(COUNT(sa.id), 0) AS visits,
			COALESCE(SUM(CASE WHEN sa.status = 'redirected' THEN 1 ELSE 0 END), 0) AS redirects
		FROM page p
		LEFT JOIN shorty_accesses sa ON sa.shorty_id = p.id
		GROUP BY %s
		ORDER BY p.created_at DESC`,
		pageFieldList, whereClause, selectFieldList, strings.Join(groupBy, ", "))

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
